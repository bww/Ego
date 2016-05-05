
SRC	= $(shell find . -name \*.go -print)

export PROJECT = $(PWD)

all: deps $(SRC)
	go test -test.v

deps:
