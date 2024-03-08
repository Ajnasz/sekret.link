BINARY_NAME=sekret.link
VERSION=$(shell git describe --tags)
BUILD=$(shell date +%FT%T%z)
BUILD_ARGS=-trimpath -ldflags '-w -s'

.PHONY: build clean linux

all: clean linux

run:
	@cd cmd/sekret.link && POSTGRES_URL="postgres://postgres:password@localhost:5432/sekret_link_test?sslmode=disable" go run . -webExternalURL=/api

build/${BINARY_NAME}.linux.amd64:
	cd cmd/sekret.link && GOARCH=amd64 GOOS=linux go build ${BUILD_ARGS} -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}" -o ../../$@

linux: build/${BINARY_NAME}.linux.amd64

clean:
	rm -f build/${BINARY_NAME}.*.*

.PHONY: test
test:
	go test -v ./...

.PHONY: curl
curl:
	curl -v --data-binary @go.mod localhost:8080/api/ | xargs -I {} curl localhost:8080{}
