GO_EXECUTABLE ?= go

MODULE = github.com/musenwill/raftdemo

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))

COMMIT := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
VERSION := $(shell git describe --always --tags --abbrev=0)
TIME := $(shell date +%FT%T%z)

LDFLAGS := $(LDFLAGS) -X ${MODULE}/common.Version=${VERSION}
LDFLAGS := $(LDFLAGS) -X ${MODULE}/common.BuildTime=${TIME}
LDFLAGS := $(LDFLAGS) -X ${MODULE}/common.Branch=${BRANCH}
LDFLAGS := $(LDFLAGS) -X ${MODULE}/common.Commit=${COMMIT}

UNAME = $(shell uname)
ifeq (${UNAME}, Darwin)
	os=darwin
else
	os=linux
endif

build: raft shell

raft: check
	${GO_EXECUTABLE} build -ldflags '${LDFLAGS}' -o bin/rafts cmd/main/main.go

shell: check
	${GO_EXECUTABLE} build -ldflags '${LDFLAGS}' -o bin/raftc shell/main/main.go

check:
	golint ./... | grep -v "exported" | exit 0
	go vet ./...
	gofmt -d -s `find . -name "*.go" -type f`
	go test ./...

clean:
	rm -rf dist bin


.PHONY: build clean check
