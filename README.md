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

