ifndef GOPATH
	GOPATH := $(HOME)/go
endif

ifndef GOOS
	GOOS := linux
endif

ifndef GO111MODULE
	GO111MODULE := on
endif

all: build

build:
	@GOPATH=$(GOPATH) CGO_ENABLED=0 go build
clean:
	rm -f rdr-emailer
