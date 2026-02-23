# Makefile for cbq (Clipboard Queue)

BINARY=cbq
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
LDFLAGS=-ldflags "-X main.version=${VERSION}"



.PHONY: all build test clean run

all: test build

build:
	go build ${LDFLAGS} -o ${BINARY} main.go

test:
	go test ./... -cover

clean:
	rm -f ${BINARY}
	rm -rf dist/

run: build
	./${BINARY}

install:
	go install ${LDFLAGS}

release-dry-run:
	goreleaser release --snapshot --clean --skip=publish

release:
	goreleaser release --clean
