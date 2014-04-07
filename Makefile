
SOURCES=\
	ego.go \
	ego_test.go \
	scanner.go \
	parser.go

all: deps $(SOURCES)
	go test

deps:

