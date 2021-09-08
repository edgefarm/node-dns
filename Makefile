NAME = node-dns
BIN_DIR ?= bin
VERSION ?= $(shell git describe --match=NeVeRmAtCh --always --abbrev=40 --dirty)
GO_LDFLAGS = -tags 'netgo osusergo static_build' -ldflags "-X github.com/siredmar/node-dns/cmd/node-dns/cmd.version=$(VERSION)"

all: test build

build: amd64 arm64

amd64:
	GOOS=linux GOARCH=amd64 go build $(GO_LDFLAGS) -o ${BIN_DIR}/${NAME}-amd64 cmd/node-dns/main.go

arm64:	
	GOOS=linux GOARCH=arm64 go build $(GO_LDFLAGS) -o ${BIN_DIR}/${NAME}-arm64 cmd/node-dns/main.go

test:
	go test ./...

clean: clean-amd64 clean-arm64

clean-amd64:
	rm -rf ${BIN_DIR}/${NAME}-amd64 

clean-arm64:
	rm -rf ${BIN_DIR}/${NAME}-arm64 
	
.PHONY: test clean clean-amd64 clean-arm64 build amd64 arm64
