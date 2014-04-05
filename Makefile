
SOURCES=\
	ego.go \
	ego_test.go \
	scanner.go

all: deps $(SOURCES)
	go test

deps:

