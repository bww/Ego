
SOURCES	= $(shell find . -name \*.go -print)
GOPATH := $(shell cd ../.. && pwd):$(GOPATH)

export PROJECT = $(PWD)

all: deps $(SOURCES)
	go test -test.v

deps:
