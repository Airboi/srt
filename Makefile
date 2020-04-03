SSHHost = 127.0.0.1
SSHPort = 2222
SSHUser = tom
SSHPwd = fuckingkey
SocksUser = foo
SocksPwd = bar
Tag = mail-server
SERVER_SOURCE = ./server/...
CLIENT_SOURCE = ./client/...
LDFLAGS="-s -w -X main.SSHHost=$(SSHHost) -X main.SSHPort=$(SSHPort) -X main.SSHUser=$(SSHUser) -X main.SSHPwd=$(SSHPwd) -X main.SocksUser=$(SocksUser) -X main.SocksPwd=$(SocksPwd) -X main.Tag=$(Tag)"
GCFLAGS="all=-trimpath=$GOPATH"

CLIENT_BINARY=client
SERVER_BINARY=server

TAGS=release

OSARCH = "linux/amd64 linux/386 linux/arm windows/amd64 windows/386 darwin/amd64 darwin/386"

.DEFAULT: help

help: ## Show Help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: build-client build-server ## Build all

dep: ## Build dep
	go get github.com/mitchellh/gox

build-client: ## Build client
	@echo "Building shell"
	gox -osarch=$(OSARCH) -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -output "release/$(CLIENT_BINARY)_{{.OS}}_{{.Arch}}" $(CLIENT_SOURCE)

build-server: ## Build server
	@echo "Building server"
	gox -osarch=$(OSARCH) -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -output "release/$(SERVER_BINARY)_{{.OS}}_{{.Arch}}" $(SERVER_SOURCE)


clean: ## Remove all the generated binaries
	rm -f release/*