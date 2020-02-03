// +build postgres

package backends

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

var postgres Postgres

var pgStrictAcl = "test/topic/1"

//Insert a user to test auth
var pgUsername = "test"
var pgUserPass = "testpw"

//Hash generated by the pw utility
var pgUserPassHash = "PBKDF2$sha512$100000$os24lcPr9cJt2QDVWssblQ==$BK1BQ2wbwU1zNxv3Ml3wLuu5//hPop3/LvaPYjjCwdBvnpwusnukJPpcXQzyyjOlZdieXTx6sXAcX4WnZRZZnw=="

func init() {
	//Initialize Postgres.
	authOpts := make(map[string]string)
	authOpts["pg_host"] = "localhost"
	authOpts["pg_port"] = "5432"
	authOpts["pg_dbname"] = "go_auth_test"
	authOpts["pg_user"] = "go_auth_test"
	authOpts["pg_password"] = "go_auth_test"
	authOpts["pg_userquery"] = "SELECT password_hash FROM test_user WHERE username = $1 limit 1"
	authOpts["pg_superquery"] = "select count(*) from test_user where username = $1 and is_admin = true"
	authOpts["pg_aclquery"] = "SELECT test_acl.topic FROM test_acl, test_user WHERE test_user.username = $1 AND test_acl.test_user_id = test_user.id AND (rw = $2 or rw = 3)"
	var err error
	postgres, err = NewPostgres(authOpts, log.DebugLevel)
	if err != nil {
		log.Fatalf("Postgres error: %s", err)
	}
	//Empty db
	postgres.DB.MustExec("delete from test_user where 1 = 1")
	postgres.DB.MustExec("delete from test_acl where 1 = 1")

	userID := 0

	insertQuery := "INSERT INTO test_user(username, password_hash, is_admin) values($1, $2, $3) returning id"
	postgres.DB.Get(&userID, insertQuery, pgUsername, pgUserPassHash, true)

	aclID := 0
	//Insert acls
	aclQuery := "INSERT INTO test_acl(test_user_id, topic, rw) values($1, $2, $3) returning id"
	postgres.DB.Get(&aclID, aclQuery, userID, strictAcl, 1)

}

func BenchmarkPostgresUser(b *testing.B) {
	log.Printf("postgres: %v", postgres)
	for n := 0; n < b.N; n++ {
		postgres.GetUser(pgUsername, pgUserPass)
	}
}

func BenchmarkPostgresSuperser(b *testing.B) {
	for n := 0; n < b.N; n++ {
		postgres.GetSuperuser(pgUsername)
	}
}

func BenchmarkPostgresStrictAcl(b *testing.B) {
	for n := 0; n < b.N; n++ {
		postgres.CheckAcl(pgUsername, "test/topic/1", "test_id", 1)
	}
}

func BenchmarkPostgresSingleLevelAcl(b *testing.B) {
	for n := 0; n < b.N; n++ {
		postgres.CheckAcl(pgUsername, "test/topic/+", "test_id", 1)
	}
}

func BenchmarkPostgresHierarchyAcl(b *testing.B) {
	for n := 0; n < b.N; n++ {
		postgres.CheckAcl(pgUsername, "test/#", "test_id", 1)
	}
}
