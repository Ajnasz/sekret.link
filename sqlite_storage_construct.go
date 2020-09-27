// +build sqlite,!est

package main

func newStorage() EntryStorage {
	return newSQLiteStorage(getConnectionString(sqliteDB, "SQLITE_DB"))
}
