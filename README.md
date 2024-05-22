# Sekret.link

Service to share notes, passwords securely.

## Build

```sh
go get
go build .
```

## Start

```sh
./sekret.link [option]...
```

## Options

`webExternalURL` when using a reverse proxy to pass requests to sekret.link and you prefix the path with something, this sould be set to the same prefix, eg.: `https://sekret.link/api`
`postgresDB` mongodb connection string
`expireSeconds` default expire time, while a secret is walid
`maxExpireSeconds` the longest time a secret can be stored
`maxDataSize` maximum size of secret in bytes
`version` print the version


Postgres URL can be set from env var:

```
POSTGRES_URL="user=sekret_link password=password host=localhost dbname=sekret_link sslmode=disable"
```

## Example usage

### Start the server

```sh
docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=sekret_link_test -d postgres:13-alpine
# in cmd/sekret.link folder
POSTGRES_URL="postgres://postgres:password@localhost:5432/sekret_link_test?sslmode=disable" go run . -webExternalURL=/api
```

### Send and receive data
```sh
curl -v -H 'content-type: text/plain' --data-binary @go.mod localhost:8080/api/ | xargs -I {} curl localhost:8080{}
```

```sh
curl -v -F 'secret=@README.md;type=text/x-markdown' localhost:8080/api/ | xargs -I {} curl -v localhost:8080{}
```
