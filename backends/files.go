package backends

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/iegomez/mosquitto-go-auth/common"
)

// saltSize defines the salt size
const saltSize = 16

// HashIterations defines the number of hash iterations.
var HashIterations = 100000

//FileUer keeps a user password and acl records.
type FileUser struct {
	Password string
}

//AclRecord holds a topic and access privileges.
type AclRecord struct {
	Topic string
	Acc   byte //None 0x00, Read 0x01, Write 0x02, ReadWrite: Read | Write : 0x03
}

//FileBE holds paths to files, list of file users and general (no user or pattern) acl records.
type Files struct {
	PasswordPath   string
	AclPath        string
	CheckAcls      bool
	Users          map[string]*FileUser //Users keeps a registry of username/FileUser pairs, holding a user's password and Acl records.
	UserAclRecords map[string][]AclRecord
	AclRecords     []AclRecord
}

//NewFiles initializes a files backend.
func NewFiles(authOpts map[string]string, logLevel log.Level) (Files, error) {

	log.SetLevel(logLevel)

	var files = Files{
		PasswordPath:   "",
		AclPath:        "",
		CheckAcls:      false,
		Users:          make(map[string]*FileUser),
		UserAclRecords: make(map[string][]AclRecord),
		AclRecords:     make([]AclRecord, 0, 0),
	}

	if passwordPath, ok := authOpts["password_path"]; ok {
		files.PasswordPath = passwordPath
	} else {
		return files, errors.New("Files backend error: no password path given.\n")
	}

	if aclPath, ok := authOpts["acl_path"]; ok {
		files.AclPath = aclPath
		files.CheckAcls = true
	} else {
		files.CheckAcls = false
		log.Info("Acls won't be checked.\n")
	}

	//Now initialize FileUsers by reading from password and acl files.
	uCount, uErr := files.readPasswords()
	if uErr != nil {
		return files, errors.Errorf("Fatal: %s\n", uErr)
	} else {
		log.Infof("Got %d users from passwords file.\n", uCount)
	}

	//Only read acls if path was given.
	if files.CheckAcls {
		aclCount, aclErr := files.readAcls()
		if aclErr != nil {
			return files, errors.Errorf("Fatal: %s\n", aclErr)
		} else {
			log.Infof("Got %d lines from acl file.\n", aclCount)
		}
	}

	return files, nil

}

//ReadPasswords read file and populates FileUsers. Return amount of users seen and possile error.
func (o *Files) readPasswords() (int, error) {

	users := make(map[string]*FileUser)
	usersCount := 0

	file, fErr := os.Open(o.PasswordPath)
	defer file.Close()
	if fErr != nil {
		return usersCount, fmt.Errorf("Files backend error: couldn't open passwords file: %s\n", fErr)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	index := 0
	//Read line by line
	for scanner.Scan() {
		index++

		//Check comment or empty line to skip them.
		if checkCommentOrEmpty(scanner.Text()) {
			continue
		}

		lineArr := strings.Split(scanner.Text(), ":")
		if len(lineArr) != 2 {
			log.Errorf("Read passwords error: line %d is not well formatted.\n", index)
			continue
		}
		//Create user if it doesn't exist and save password; override password if user existed.
		var fileUser *FileUser
		var ok bool
		fileUser, ok = users[lineArr[0]]
		if ok {
			fileUser.Password = lineArr[1]
		} else {
			usersCount++
			fileUser = &FileUser{
				Password: lineArr[1],
			}
			users[lineArr[0]] = fileUser
		}
	}
	o.Users = users
	log.Debugf("users from file: ")
	for k := range o.Users {
		log.Debugf(" %s", k)
	}

	return usersCount, nil

}

//ReadAcls reads the Acl file and associates them to existing users. It omits any non existing users.
func (o *Files) readAcls() (int, error) {
	aclRecords := make([]AclRecord, 0, 0)
	userAclRecords := make(map[string][]AclRecord)
	linesCount := 0

	//Set currentUser as empty string
	currentUser := ""

	file, fErr := os.Open(o.AclPath)
	defer file.Close()
	if fErr != nil {
		return linesCount, errors.Errorf("Files backend error: couldn't open acl file: %s\n", fErr)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	index := 0

	for scanner.Scan() {
		index++
		line := scanner.Text()

		//Check comment or empty line to skip them.
		if checkCommentOrEmpty(scanner.Text()) {
			continue
		}

		//If we see a user line, change the current user.
		if strings.Contains(line, "user") {
			//Try to get username
			lineArr := strings.Fields(line)

			//Check format
			if len(lineArr) == 2 && lineArr[0] == "user" {
				_, ok := o.Users[lineArr[1]]

				//Check that user exists
				if !ok {
					return 0, errors.Errorf("Files backend error: user %s does not exist for acl at line %d\n", lineArr[1], index)
				}

				currentUser = lineArr[1]

			} else {
				return 0, errors.Errorf("Files backend error: wrong acl format at line %d\n", index)
			}
		} else if strings.Contains(line, "topic") {

			//Split and check for read, write or empty (readwwrite) privileges.
			lineArr := strings.Fields(line)

			if (len(lineArr) == 2 || len(lineArr) == 3) && lineArr[0] == "topic" {

				var aclRecord = AclRecord{
					Topic: "",
					Acc:   MOSQ_ACL_NONE,
				}

				//If len is 2, then we assume ReadWrite privileges.
				if len(lineArr) == 2 {
					aclRecord.Topic = lineArr[1]
					aclRecord.Acc = MOSQ_ACL_READWRITE
				} else {
					aclRecord.Topic = lineArr[2]
					if lineArr[1] == "read" {
						aclRecord.Acc = MOSQ_ACL_READ
					} else if lineArr[1] == "write" {
						aclRecord.Acc = MOSQ_ACL_WRITE
					} else if lineArr[1] == "readwrite" {
						aclRecord.Acc = MOSQ_ACL_READWRITE
					} else if lineArr[1] == "subscribe" {
						aclRecord.Acc = MOSQ_ACL_SUBSCRIBE
					} else {
						return 0, errors.Errorf("Files backend error: wrong acl format at line %d\n", index)
					}
				}

				//Append to user or general depending on currentUser.
				if currentUser != "" {
					//					userRecords, ok := userAclRecords[currentUser]
					//					if !ok {
					//						userRecords = make([]AclRecord, 0, 0)
					//					}
					userAclRecords[currentUser] = append(userAclRecords[currentUser], aclRecord)
				} else {
					aclRecords = append(aclRecords, aclRecord)
				}

				linesCount++

			} else {
				return 0, errors.Errorf("Files backend error: wrong acl format at line %d\n", index)
			}

		} else if strings.Contains(line, "pattern") {

			//Split and check for read, write or empty (readwwrite) privileges.
			lineArr := strings.Fields(line)

			if (len(lineArr) == 2 || len(lineArr) == 3) && lineArr[0] == "pattern" {

				var aclRecord = AclRecord{
					Topic: "",
					Acc:   MOSQ_ACL_NONE,
				}

				//If len is 2, then we assume ReadWrite privileges.
				if len(lineArr) == 2 {
					aclRecord.Topic = lineArr[1]
					aclRecord.Acc = MOSQ_ACL_READWRITE
				} else {
					aclRecord.Topic = lineArr[2]
					if lineArr[1] == "read" {
						aclRecord.Acc = MOSQ_ACL_READ
					} else if lineArr[1] == "write" {
						aclRecord.Acc = MOSQ_ACL_WRITE
					} else if lineArr[1] == "readwrite" {
						aclRecord.Acc = MOSQ_ACL_READWRITE
					} else if lineArr[1] == "subscribe" {
						aclRecord.Acc = MOSQ_ACL_SUBSCRIBE
					} else {
						return 0, errors.Errorf("Files backend error: wrong acl format at line %d\n", index)
					}
				}

				//Append to general acls.
				aclRecords = append(aclRecords, aclRecord)

				linesCount++

			} else {
				return 0, errors.Errorf("Files backend error: wrong acl format at line %d\n", index)
			}

		}
	}
	o.UserAclRecords = userAclRecords
	o.AclRecords = aclRecords
	return linesCount, nil

}

func checkCommentOrEmpty(line string) bool {
	if len(strings.Replace(line, " ", "", -1)) == 0 || line[0:1] == "#" {
		return true
	}
	return false
}

//GetUser checks that user exists and password is correct.
func (o Files) GetUser(username, password string) bool {

	fileUser, ok := o.Users[username]
	if !ok {
		return false
	}

	if common.HashCompare(password, fileUser.Password) {
		return true
	}

	log.Infof("[files] wrong password for user %s\n", username)

	return false

}

//GetSuperuser returns false for files backend.
func (o Files) GetSuperuser(username string) bool {
	return false
}

//CheckAcl checks that the topic may be read/written by the given user/clientid.
func (o Files) CheckAcl(username, topic, clientid string, acc int32) bool {
	//If there are no acls, all access is allowed.
	if !o.CheckAcls {
		return true
	}

	fileUserRecords, ok := o.UserAclRecords[username]

	//If user exists, check against his acls and common ones. If not, check against common acls only.
	if ok {
		for _, aclRecord := range fileUserRecords {
			log.Debugf("fileUserRecord = %s", aclRecord.Topic)
			if common.TopicsMatch(aclRecord.Topic, topic) && (acc == int32(aclRecord.Acc) || int32(aclRecord.Acc) == MOSQ_ACL_READWRITE || (acc == MOSQ_ACL_SUBSCRIBE && topic != "#" && (int32(aclRecord.Acc) == MOSQ_ACL_READ || int32(aclRecord.Acc) == MOSQ_ACL_SUBSCRIBE))) {
				return true
			}
		}
	}
	for _, aclRecord := range o.AclRecords {
		//Replace all occurrences of %c for clientid and %u for username
		aclTopic := strings.Replace(aclRecord.Topic, "%c", clientid, -1)
		aclTopic = strings.Replace(aclTopic, "%u", username, -1)
		log.Debugf("acltopic = %s (%s)", aclTopic, aclRecord.Topic)
		if common.TopicsMatch(aclTopic, topic) && (acc == int32(aclRecord.Acc) || int32(aclRecord.Acc) == MOSQ_ACL_READWRITE || (acc == MOSQ_ACL_SUBSCRIBE && topic != "#" && (int32(aclRecord.Acc) == MOSQ_ACL_READ || int32(aclRecord.Acc) == MOSQ_ACL_SUBSCRIBE))) {
			return true
		}
	}

	return false

}

//GetName returns the backend's name
func (o Files) GetName() string {
	return "Files"
}

//Halt does nothing for files as there's no cleanup needed.
func (o Files) Halt() {
	//Do nothing
}

func (o Files) Reload() {
	log.Info("Read passwords")
	o.readPasswords()
	log.Info("Read acls")
	o.readAcls()
}
