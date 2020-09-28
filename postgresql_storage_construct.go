// +build postgres,!test

package main

import "github.com/Ajnasz/sekret.link/storage"

func newStorage() storage.EntryStorage {
	return newPostgresqlStorage(getConnectionString(postgresDB, "POSTGRES_URL"))
}
