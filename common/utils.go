package common

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
    "golang.org/x/crypto/bcrypt" 

	"github.com/jmoiron/sqlx"
)

// OpenDatabase opens the database and performs a ping to make sure the
// database is up.
// Taken from brocaar's lora-app-server: https://github.com/brocaar/lora-app-server
func OpenDatabase(dsn, engine string) (*sqlx.DB, error) {

	db, err := sqlx.Open(engine, dsn)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %s", err)
	}

	for {
		if err = db.Ping(); err != nil {
			log.Printf("ping database error, will retry in 2s: %s", err)
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	return db, nil
}

func TopicsMatch(savedTopic, givenTopic string) bool {
	return givenTopic == savedTopic || match(strings.Split(savedTopic, "/"), strings.Split(givenTopic, "/"))
}

func match(route []string, topic []string) bool {
	if len(route) == 0 {
		if len(topic) == 0 {
			return true
		}
		return false
	}

	if len(topic) == 0 {
		if route[0] == "#" {
			return true
		}
		return false
	}

	if route[0] == "#" {
		return true
	}

	if (route[0] == "+") || (route[0] == topic[0]) {
		return match(route[1:], topic[1:])
	}

	return false
}

func Hash(password string, cost int) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func HashCompare(password string, passwordHash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
    return err == nil
}
