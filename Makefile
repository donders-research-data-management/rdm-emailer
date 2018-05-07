GOPATH := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
PREFIX ?= "/opt/project"

all: build

external:
	@GOPATH=$(GOPATH)  go get ./...

build: external
	@GOPATH=$(GOPATH) CGO_ENABLED=0  go install ./...

doc:
	@GOPATH=$(GOPATH)  godoc -http=:6060

test: external
	@GOPATH=$(GOPATH)  GOCACHE=off go test -v dccn.nl/project/...

install: build
	@install -D $(GOPATH)/bin/* $(PREFIX)/bin

clean:
	@rm -rf bin
	@rm -rf pkg
