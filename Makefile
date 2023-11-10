BINARY_NAME=sekret.link
VERSION=$(shell git describe --tags)
BUILD=$(shell date +%FT%T%z)
BUILD_ARGS=-trimpath -ldflags '-w -s'

.PHONY: build clean linux

all: clean linux

run:
	@go run .

build/${BINARY_NAME}.linux.amd64:
	cd cmd/sekret.link && GOARCH=amd64 GOOS=linux go build ${BUILD_ARGS} -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}" -o ../../$@

linux: build/${BINARY_NAME}.linux.amd64

clean:
	rm -f build/${BINARY_NAME}.*.*

.PHONY: test
test:
	go test -v ./...
