BINARY_NAME=sekret.link
VERSION=$(shell git describe --tags)
BUILD=$(shell date +%FT%T%z)
BUILD_ARGS=-trimpath -ldflags '-w -s'

.PHONY: build clean linux

all: clean linux

run:
	@cd cmd/sekret.link && POSTGRES_URL="postgres://postgres:password@localhost:5432/sekret_link_test?sslmode=disable" go run . -webExternalURL=/api -base62

build/${BINARY_NAME}.linux.amd64:
	cd cmd/sekret.link && GOARCH=amd64 GOOS=linux go build ${BUILD_ARGS} -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}" -o ../../$@

linux: build/${BINARY_NAME}.linux.amd64

clean:
	rm -f build/${BINARY_NAME}.*.*

dbcreate:
	@cd cmd/prepare && go run .

dbcreate-test:
	@cd cmd/prepare && go run . -postgresDB "postgres://postgres:password@localhost:5432/sekret_link_test?sslmode=disable"

.PHONY: test
test: dbcreate-test
	go test ./... -count 1

.PHONY: curl curl-bad
curl:
	curl -v --data-binary @go.mod localhost:8080/api/ | xargs -I {} curl localhost:8080{}
curl-bad:
	curl -v localhost:8080/api/57c04c70-dd58-11ee-98fc-ebbaf68907f4/8bc419a2de0ccf0b165cd978f8894b77403a2f06019916af0bf48bcade88f518

.PHONY: hurl
hurl:
	@hurl --verbose --error-format=long --variable api_host='http://localhost:8080' hurl/*.hurl

.PHONY: coverage clean-cover
clean-cover:
	rm -f cover.out cover.html

coverage: cover.out cover.html

cover.out:
	go test ./... -coverprofile cover.out

cover.html: cover.out
	go tool cover -html=cover.out -o cover.html

