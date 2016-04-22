
# make sure the vendor experiment is enabled
export GO15VENDOREXPERIMENT := 1
# project path
export PROJECT = $(PWD)

# build paths
OUTPUT_DIR=$(PWD)/target
TARGET_DIR=$(OUTPUT_DIR)
BUILD_DIR=$(TARGET_DIR)/product

# product name
PRODUCT_NAME=egoc

# packaging
PRODUCT=$(BUILD_DIR)/bin/$(PRODUCT_NAME)
ARCHIVE_NAME=$(PRODUCT_NAME).tar.gz
ARCHIVE=$(TARGET_DIR)/$(ARCHIVE_NAME)

# source paths
SRC		  = $(shell find . -name \*.go -print)
LIB		  = $(shell find .. -name \*.go -print)
GOPATH := $(shell cd ../../.. && pwd):$(GOPATH)

.PHONY: all deps build clean

all: build

deps:

$(PRODUCT): $(SRC) $(LIB)
	mkdir -p $(BUILD_DIR)/bin
	go build -o $@

build: deps $(PRODUCT)

clean:
	rm -rf $(OUTPUT_DIR)
