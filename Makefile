
SOURCES=\
	ego.go \
	ego_test.go \
	scanner.go \
	parser.go \
	program.go

all: deps $(SOURCES)
	go test

deps:

