
SRC	= $(shell find . -name \*.go -print)

export PROJECT = $(PWD)

.PHONY: all deps

all: deps $(SRC)
	go test -test.v

deps:
