BINARY_NAME=swfbq
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Detect the operating system and architecture
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

.PHONY: all build clean test

all: clean build

build:
	@echo "Building for ${GOOS}/${GOARCH}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME}-${GOOS}-${GOARCH}

build-all:
	GOOS=linux GOARCH=amd64 make build
	GOOS=linux GOARCH=arm64 make build
	GOOS=darwin GOARCH=amd64 make build
	GOOS=darwin GOARCH=arm64 make build

clean:
	@echo "Cleaning..."
	@rm -rf bin/

test:
	go test -v ./...

.DEFAULT_GOAL := build
