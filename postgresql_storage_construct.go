package main

import (
	"database/sql"
	"log"
)

type dbExec func(*sql.DB) error

func addExtension(db *sql.DB) error {
	q, err := db.Prepare("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

	if err != nil {
		return err
	}
	_, err = q.Exec()

	return err
}

func createTable(db *sql.DB) error {
	q, err := db.Prepare("CREATE TABLE IF NOT EXISTS entries (uuid uuid PRIMARY KEY, data BYTEA, remaining_reads SMALLINT DEFAULT 1, delete_key CHAR(256) NOT NULL, created TIMESTAMPTZ, accessed TIMESTAMPTZ, expire TIMESTAMPTZ);")

	if err != nil {
		return err
	}
	_, err = q.Exec()

	return err
}

func addRemainingRead(db *sql.DB) error {
	alterTable, err := db.Prepare("ALTER TABLE entries ADD COLUMN IF NOT EXISTS remaining_reads SMALLINT DEFAULT 1;")

	if err != nil {
		return err
	}

	_, err = alterTable.Exec()

	return err
}

func addDeleteKey(db *sql.DB) error {
	alterTable, err := db.Prepare("ALTER TABLE entries ADD COLUMN IF NOT EXISTS delete_key CHAR(256) NOT NULL;")

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
		log.Fatal("DB ping failed", err)
	}

	for _, f := range []dbExec{addExtension, createTable, addRemainingRead, addDeleteKey} {
		err = f(db)
		if err != nil {
			defer db.Close()
			log.Fatal("Migrate db failed", err)
		}
	}

	return &postgresqlStorage{db}
}

func newStorage() verifyStorage {
	return newPostgresqlStorage(getConnectionString(postgresDB, "POSTGRES_URL"))
}
