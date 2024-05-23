package migrate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Ajnasz/sekret.link/internal/key"
)

type EntryMigration struct{}

func NewEntryMigration() *EntryMigration {
	return &EntryMigration{}
}

func (*EntryMigration) Create(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS entries (
		uuid uuid PRIMARY KEY,
		data BYTEA,
		remaining_reads SMALLINT DEFAULT 1,
		delete_key CHAR(256) NOT NULL,
		created TIMESTAMPTZ DEFAULT NOW(),
		accessed TIMESTAMPTZ,
		expire TIMESTAMPTZ DEFAULT NULL
	);`)

	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (e *EntryMigration) Alter(ctx context.Context, tx *sql.Tx) error {
	if err := e.addRemainingRead(ctx, tx); err != nil {
		return err
	}
	if err := e.addDeleteKey(ctx, tx); err != nil {
		return err
	}

	if err := e.addContentType(ctx, tx); err != nil {
		return err
	}

	if err := e.dropKeyFields(ctx, tx); err != nil {
		return err
	}

	return nil
}

func (*EntryMigration) addRemainingRead(ctx context.Context, tx *sql.Tx) error {
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

func (*EntryMigration) addDeleteKey(ctx context.Context, tx *sql.Tx) error {
	alterTable, err := tx.PrepareContext(ctx, "ALTER TABLE entries ADD COLUMN IF NOT EXISTS delete_key CHAR(256);")

	if err != nil {
		return fmt.Errorf("failed to add delete_key column: %w", err)
	}

	_, err = alterTable.ExecContext(ctx)

	if err != nil {
		return fmt.Errorf("failed to add delete_key column: %w", err)
	}

	rows, err := tx.QueryContext(ctx, "SELECT uuid FROM entries WHERE delete_key IS NULL;")
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

		deleteKey := k.String()

		_, err = tx.ExecContext(ctx, "UPDATE entries SET delete_key=$2 WHERE uuid=$1", UUID, deleteKey)
		if err != nil {
			return fmt.Errorf("failed to update delete_key: %w", err)
		}
	}
	_, err = tx.ExecContext(ctx, "ALTER TABLE entries ALTER COLUMN delete_key SET NOT NULL;")
	if err != nil {
		return fmt.Errorf("failed to alter delete_key column: %w", err)
	}

	return nil
}

func (e *EntryMigration) addContentType(ctx context.Context, tx *sql.Tx) error {
	alterTable, err := tx.PrepareContext(ctx, "ALTER TABLE entries ADD COLUMN IF NOT EXISTS content_type VARCHAR(256) NOT NULL DEFAULT '';")

	if err != nil {
		return fmt.Errorf("failed to add delete_key column: %w", err)
	}

	_, err = alterTable.Exec()

	if err != nil {
		return fmt.Errorf("failed to add remaining_reads column: %w", err)
	}

	return nil
}

func (e *EntryMigration) dropKeyFields(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "ALTER TABLE entries DROP COLUMN IF EXISTS remaining_reads;")
	if err != nil {
		return fmt.Errorf("failed to drop remaining_reads column: %w", err)
	}

	_, err = tx.ExecContext(ctx, "ALTER TABLE entries DROP COLUMN IF EXISTS expire;")
	if err != nil {
		return fmt.Errorf("failed to drop expire column: %w", err)
	}

	return nil
}
