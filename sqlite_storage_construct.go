// +build sqlite,!est

package main

import "github.com/Ajnasz/sekret.link/storage"

func newStorage() storage.EntryStorage {
	return newSQLiteStorage(getConnectionString(sqliteDB, "SQLITE_DB"))
}
