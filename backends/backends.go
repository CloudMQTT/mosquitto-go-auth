package backends

import (
	log "github.com/sirupsen/logrus"
)

type Backend interface {
	GetUser(username, password string) bool
	GetSuperuser(username string) bool
	CheckAcl(username, topic, clientId string, acc int32) bool
	GetName() string
	Halt()
	Reload()
}
type createFunc func(authOpts map[string]string, logLevel log.Level) (Backend, error)

var RegisteredBackends = make(map[string]createFunc)
