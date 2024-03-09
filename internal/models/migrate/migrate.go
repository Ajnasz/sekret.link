package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/Ajnasz/sekret.link/internal/key"
)

type dbExec func(context.Context, *sql.Tx) error

func createTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS entries (
		uuid uuid PRIMARY KEY,
		data BYTEA,
		remaining_reads SMALLINT DEFAULT 1,
		delete_key CHAR(256) NOT NULL,
		created TIMESTAMPTZ,
		accessed TIMESTAMPTZ,
		expire TIMESTAMPTZ
	);`)

	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func addRemainingRead(ctx context.Context, tx *sql.Tx) error {
	alterTable, err := tx.PrepareContext(ctx, "ALTER TABLE entries ADD COLUMN IF NOT EXISTS remaining_reads SMALLINT DEFAULT 1;")

	if err != nil {
		return err
	}

	_, err = alterTable.Exec()

	if err != nil {
		return fmt.Errorf("failed to add remaining_reads column: %w", err)
	}

	return nil
}

func addDeleteKey(ctx context.Context, db *sql.Tx) error {
	alterTable, err := db.PrepareContext(ctx, "ALTER TABLE entries ADD COLUMN IF NOT EXISTS delete_key CHAR(256);")

	if err != nil {
		return fmt.Errorf("failed to add delete_key column: %w", err)
	}

	_, err = alterTable.ExecContext(ctx)

	if err != nil {
		return fmt.Errorf("failed to add delete_key column: %w", err)
	}

	rows, err := db.QueryContext(ctx, "SELECT uuid FROM entries WHERE delete_key IS NULL;")
	if err != nil {
		return err
	}

	for rows.Next() {
		var UUID string
		if err := rows.Scan(&UUID); err != nil {
			return fmt.Errorf("failed to scan UUID: %w", err)
		}

		k, err := key.NewGeneratedKey()
		if err != nil {
			return err
		}

		deleteKey := k.ToHex()

		_, err = db.ExecContext(ctx, "UPDATE entries SET delete_key=$2 WHERE uuid=$1", UUID, deleteKey)
		if err != nil {
			return fmt.Errorf("failed to update delete_key: %w", err)
		}
	}
	_, err = db.ExecContext(ctx, "ALTER TABLE entries ALTER COLUMN delete_key SET NOT NULL;")
	if err != nil {
		return fmt.Errorf("failed to alter delete_key column: %w", err)
	}

	return nil
}

func prepareDatabase(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	actions := []dbExec{
		createTable,
		addRemainingRead,
		addDeleteKey,
	}

	for _, action := range actions {
		if err := action(ctx, tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute action: %w", err)
		}
	}

	return tx.Commit()

}

var once = sync.Once{}

func PrepareDatabase(ctx context.Context, db *sql.DB) error {
	var err error
	once.Do(func() {
		err = prepareDatabase(ctx, db)
	})
	return err
}
