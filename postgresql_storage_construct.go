package main

import (
	"database/sql"
	"log"

	"github.com/Ajnasz/sekret.link/storage"
)

type dbExec func(*sql.DB) error

func createTable(db *sql.DB) error {
	createTable, err := db.Prepare("CREATE TABLE IF NOT EXISTS entries (uuid uuid PRIMARY KEY, data BYTEA, remaining_reads SMALLINT DEFAULT 1, created TIMESTAMPTZ, accessed TIMESTAMPTZ, expire TIMESTAMPTZ)")

	if err != nil {
		return err
	}
	_, err = createTable.Exec()

	return err
}

func addRemainingRead(db *sql.DB) error {
	alterTable, err := db.Prepare("ALTER TABLE entries ADD COLUMN IF NOT EXISTS remaining_reads SMALLINT DEFAULT 1")

	if err != nil {
		return err
	}

	_, err = alterTable.Exec()

	return err
}

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

	for _, f := range []dbExec{createTable, addRemainingRead} {
		err = f(db)
		if err != nil {
			defer db.Close()
			log.Fatal(err)
		}
	}

	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}

	return &postgresqlStorage{db}
}

func newStorage() storage.EntryStorage {
	return newPostgresqlStorage(getConnectionString(postgresDB, "POSTGRES_URL"))
}
