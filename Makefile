BACKENDS='files http'

all:
	CGO_LDFLAGS_ALLOW="-undefined|dynamic_lookup" go build -tags=$(BACKENDS) -buildmode=c-archive go-auth.go
	CGO_LDFLAGS_ALLOW="-undefined|dynamic_lookup" go build -tags=$(BACKENDS) -buildmode=c-shared -o go-auth.so

linux:
	vagrant up || vagrant reload
	vagrant ssh -c "cd ~/go/src/github.com/CloudMQTT/mosquitto-go-auth; CGO_LDFLAGS_ALLOW='-undefined|dynamic_lookup' go build -tags=$(BACKENDS) -buildmode=c-archive go-auth.go -o go-auth-linux.a"
	vagrant ssh -c "cd ~/go/src/github.com/CloudMQTT/mosquitto-go-auth; CGO_LDFLAGS_ALLOW='-undefined|dynamic_lookup' go build -tags=$(BACKENDS) -buildmode=c-shared -o go-auth.so -o go-auth-linux.so"
	vagrant halt

requirements:
	dep ensure -v

dev-requirements:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/smartystreets/goconvey

test:
	go test ./backends -v -bench=none -count=1

benchmark:
	go test ./backends -v -bench=. -run=^a
