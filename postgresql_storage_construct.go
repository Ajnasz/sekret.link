// +build postgres,!test

package main

func newStorage() EntryStorage {
	return newPostgresqlStorage(getConnectionString(postgresDB, "POSTGRES_URL"))
}
