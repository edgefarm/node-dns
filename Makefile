NAME = edgefarm
BIN_DIR ?= ../bin
VERSION ?= $(shell git describe --match=NeVeRmAtCh --always --abbrev=40 --dirty)
GO_LDFLAGS = -tags 'netgo osusergo static_build' -ldflags "-X github.com/edgefarm/edgefarm-cli/cli/cmd.version=$(VERSION)"
GO_ARCH = amd64

all: check test build

check:
ifneq ($(shell ls openapi_dlm openapi_alm >/dev/null 2>&1; echo $$?), 0)
		@echo "generated openapi directories not found. Run 'make api'"; exit 1
endif

api:
	cd ../ && ./dobi.sh generate-client-sources

build:
	GOOS=linux GOARCH=${GO_ARCH} go build $(GO_LDFLAGS) -o ${BIN_DIR}/${NAME} main.go
	GOOS=windows GOARCH=${GO_ARCH} go build $(GO_LDFLAGS) -o ${BIN_DIR}/${NAME}.exe main.go

test:
	go test ./...

clean:
	rm -rf ${BIN_DIR}/${NAME} ${BIN_DIR}/${NAME}.exe openapi_dlm openapi_alm

.PHONY: check test clean
