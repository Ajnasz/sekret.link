package storage

import (
	"context"
	"database/sql"
	"log"

	"github.com/Ajnasz/sekret.link/key"
)

type dbExec func(*sql.DB) error

func createTable(db *sql.DB) error {
	q, err := db.Prepare(`CREATE TABLE IF NOT EXISTS
	entries (
		uuid uuid PRIMARY KEY,
		data BYTEA,
		remaining_reads SMALLINT DEFAULT 1,
		delete_key CHAR(256) NOT NULL,
		created TIMESTAMPTZ,
		accessed TIMESTAMPTZ,
		expire TIMESTAMPTZ
	);`)

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
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	alterTable, err := db.PrepareContext(ctx, "ALTER TABLE entries ADD COLUMN IF NOT EXISTS delete_key CHAR(256);")

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = alterTable.ExecContext(ctx)

	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err := db.QueryContext(ctx, "SELECT uuid FROM entries WHERE delete_key IS NULL;")
	if err != nil {
		tx.Rollback()
		return err
	}

	for rows.Next() {
		var UUID string
		if err := rows.Scan(&UUID); err != nil {
			tx.Rollback()
			return err
		}

		k, err := key.NewGeneratedKey()
		if err != nil {
			tx.Rollback()
			return err
		}

		deleteKey := k.ToHex()

		_, err = db.ExecContext(ctx, "UPDATE entries SET delete_key=$2 WHERE uuid=$1", UUID, deleteKey)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	_, err = db.ExecContext(ctx, "ALTER TABLE entries ALTER COLUMN delete_key SET NOT NULL;")
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func ConnectToPostgresql(psqlURL string) *PostgresqlStorage {
	db, err := sql.Open("postgres", psqlURL)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		defer db.Close()
		log.Fatal("DB ping failed: ", err)
	}

	for _, f := range []dbExec{createTable, addRemainingRead, addDeleteKey} {
		err = f(db)
		if err != nil {
			defer db.Close()
			log.Fatal("Migrate db failed: ", err)
		}
	}

	return NewPostgresqlStorage(db)
}

func NewStorage(connectionString string) VerifyStorage {
	return ConnectToPostgresql(connectionString)
}
