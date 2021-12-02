NAME = node-dns
BIN_DIR ?= bin
VERSION ?= $(shell git describe --match=NeVeRmAtCh --always --abbrev=40 --dirty)
GO_LDFLAGS = -tags 'netgo osusergo static_build' -ldflags "-X github.com/edgefarm/node-dns/cmd/node-dns/cmd.version=$(VERSION)"

all: test build

build:
	GOOS=linux go build $(GO_LDFLAGS) -o ${BIN_DIR}/${NAME} cmd/node-dns/main.go

test:
	go test ./...

clean:
	rm -rf ${BIN_DIR}/${NAME}

.PHONY: test clean build
