GO_EXECUTABLE ?= go
VERSION = `git describe --always --tags --abbrev=0 | tr -d "[\r\n]"`
TIME = `date +%FT%T%z`
MODULE = github.com/musenwill/raftdemo

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
BINARY := "raftdemo"

LDFLAGS= -ldflags "-X ${MODULE}/common.Version=${VERSION} -X ${MODULE}/common.BuildTime=${TIME} -X ${MODULE}/common.AppName=${BINARY}"

UNAME = $(shell uname)
ifeq (${UNAME}, Darwin)
	os=darwin
else
	os=linux
endif

build: raft shell

raft: check
	${GO_EXECUTABLE} build ${LDFLAGS} -o bin/${BINARY} cmd/main/main.go

shell: check
	${GO_EXECUTABLE} build ${LDFLAGS} -o bin/raftc shell/main/main.go

check:
	golint ./... | grep -v "exported" | exit 0
	go vet ./...
	gofmt -d -s `find . -name "*.go" -type f`
	go test ./...

clean:
	rm -rf dist bin


.PHONY: build clean check
