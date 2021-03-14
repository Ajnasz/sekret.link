package main

import (
	"database/sql"
	"log"

	"github.com/Ajnasz/sekret.link/storage"
)

func newPostgresqlStorage(psqlconn string) *postgresqlStorage {
	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}

	createTable, err := db.Prepare("CREATE TABLE IF NOT EXISTS entries (uuid uuid PRIMARY KEY, data BYTEA, created TIMESTAMPTZ, accessed TIMESTAMPTZ, expire TIMESTAMPTZ)")

	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}
	_, err = createTable.Exec()

	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}

	return &postgresqlStorage{db}
}

func newStorage() storage.EntryStorage {
	return newPostgresqlStorage(getConnectionString(postgresDB, "POSTGRES_URL"))
}
