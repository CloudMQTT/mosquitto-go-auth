package backends

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type JWT struct {
	Remote bool

	Secret string

	Method       string
	UserUri      string
	SuperuserUri string
	AclUri       string
	Hostname     string
	Port         string
	Ip           string
	WithTLS      bool
	VerifyPeer   bool
}

type Response struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

/*var allowedOpts = map[string]bool{
	"remote":        true,
	"secret":        true,
	"method":        true,
	"user_uri":      true,
	"superuser_uri": true,
	"acl_uri":       true,
	"hostname":      true,
	"port":          true,
	"ip":            true,
	"with_tls":      true,
	"verify_peer":   true,
}*/

func NewJWT(authOpts map[string]string) (JWT, error) {

	//Initialize with defaults
	var jwt = JWT{
		Remote:     false,
		Method:     "post",
		WithTLS:    false,
		VerifyPeer: false,
	}

	if remote, ok := authOpts["jwt_remote"]; ok && remote == "true" {
		jwt.Remote = true
	}

	//If remote, set remote api fields. Else, set jwt secret.

	if jwt.Remote {

		missingOpts := ""
		remoteOk := true

		if method, ok := authOpts["jwt_method"]; ok {
			jwt.Method = method
		}

		if userUri, ok := authOpts["jwt_getuser_uri"]; ok {
			jwt.UserUri = userUri
		} else {
			remoteOk = false
			missingOpts += " jwt_getuser_uri"
		}

		if superuserUri, ok := authOpts["jwt_superuser_uri"]; ok {
			jwt.SuperuserUri = superuserUri
		} else {
			remoteOk = false
			missingOpts += " jwt_superuser_uri"
		}

		if aclUri, ok := authOpts["jwt_aclcheck_uri"]; ok {
			jwt.AclUri = aclUri
		} else {
			remoteOk = false
			missingOpts += " jwt_aclcheck_uri"
		}

		if hostname, ok := authOpts["jwt_hostname"]; ok {
			jwt.Hostname = hostname
		} else {
			remoteOk = false
			missingOpts += " jwt_hostname"
		}

		if port, ok := authOpts["jwt_port"]; ok {
			jwt.Port = port
		} else {
			remoteOk = false
			missingOpts += " jwt_port"
		}

		if ip, ok := authOpts["jwt_ip"]; ok {
			jwt.Ip = ip
		} else {
			remoteOk = false
			missingOpts += " jwt_ip"
		}

		if withTLS, ok := authOpts["jwt_with_tls"]; ok && withTLS == "true" {
			jwt.WithTLS = true
		}

		if verifyPeer, ok := authOpts["jwt_verify_peer"]; ok && verifyPeer == "true" {
			jwt.VerifyPeer = true
		}

		if !remoteOk {
			log.Fatalf("JWT backend error: missing options%s.\n", missingOpts)
		}

	} else {
		if secret, ok := authOpts["jwt_secret"]; ok {
			jwt.Secret = secret
		} else {
			log.Fatal("JWT backend error: missing jwt secret.\n")
		}
	}

	return jwt, nil
}

func (o JWT) GetUser(token, password string) bool {

	if o.Remote {
		dataMap := map[string]interface{}{
			"password": token,
		}
		return httpRequest(o.Method, o.Ip, o.UserUri, token, o.WithTLS, o.VerifyPeer, dataMap, o.Port)
	}

	//If not remote, just use the secret.
	return false

}

func (o JWT) GetSuperuser(token string) bool {

	if o.Remote {
		var dataMap map[string]interface{}
		return httpRequest(o.Method, o.Ip, o.SuperuserUri, token, o.WithTLS, o.VerifyPeer, dataMap, o.Port)
	}

	//If not remote, just use the secret.
	return false

}

func (o JWT) CheckAcl(token, topic, clientid string, acc int32) bool {

	if o.Remote {
		dataMap := map[string]interface{}{
			"clientid": clientid,
			"topic":    topic,
			"acc":      acc,
		}
		return httpRequest(o.Method, o.Ip, o.AclUri, token, o.WithTLS, o.VerifyPeer, dataMap, o.Port)
	}

	//If not remote, just use the secret.
	return false

}

func httpRequest(method, host, uri, token string, withTLS, verifyPeer bool, dataMap map[string]interface{}, port string) bool {

	tlsStr := "http://"

	if withTLS {
		tlsStr = "https://"
	}

	fullUri := fmt.Sprintf("%s%s:%s%s", tlsStr, host, port, uri)

	client := &http.Client{Timeout: 5 * time.Second}

	if !verifyPeer {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	dataJson, mErr := json.Marshal(dataMap)

	if mErr != nil {
		log.Printf("marshal error: %v\n", mErr)
		return false
	}

	contentReader := bytes.NewReader(dataJson)
	req, reqErr := http.NewRequest("POST", fullUri, contentReader)

	if reqErr != nil {
		log.Printf("req error: %v\n", reqErr)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", token)

	resp, err := client.Do(req)

	if err != nil {
		log.Printf("error: %v\n", err)
		return false
	}

	body, bErr := ioutil.ReadAll(resp.Body)

	if bErr != nil {
		log.Printf("read error: %v", bErr)
		return false
	}

	response := Response{Ok: false, Error: ""}

	jErr := json.Unmarshal(body, &response)

	if jErr != nil {
		log.Printf("unmarshal error: %v", jErr)
		return false
	}

	if resp.Status != "200 OK" {
		log.Printf("error code: %v\n", err)
		return false
	} else if !response.Ok {
		log.Printf("api error: %s", response.Error)
		return false
	}

	return true

}

//GetName return the backend's name
func (o JWT) GetName() string {
	return "JWT"
}