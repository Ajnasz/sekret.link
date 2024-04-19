package migrate

import (
	"context"
	"database/sql"
	"fmt"
)

type EntryKeyMigration struct{}

func NewEntryKeyMigration() *EntryKeyMigration {
	return &EntryKeyMigration{}
}

func (e *EntryKeyMigration) Create(ctx context.Context, tx *sql.Tx) error {
	query := `
	CREATE TABLE IF NOT EXISTS entry_key (
	uuid UUID PRIMARY KEY,
	entry_uuid UUID NOT NULL,
	encrypted_key BYTEA NOT NULL,
	key_hash BYTEA NOT NULL,
	expire TIMESTAMPTZ DEFAULT NULL,
	remaining_reads SMALLINT DEFAULT NULL,
	accesed TIMESTAMPTZ DEFAULT NULL,
	created TIMESTAMPTZ,
	FOREIGN KEY (entry_uuid) REFERENCES entries(uuid) ON DELETE CASCADE
	);
`
	_, err := tx.ExecContext(ctx, query)

	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (e *EntryKeyMigration) renameAccesedToAccessed(ctx context.Context, tx *sql.Tx) error {
	// in postgresql db if it has a column named accesed rename it to accessed
	hasColumnQuery := `
	SELECT 1 as hasColumn FROM information_schema.columns
	WHERE table_name = 'entry_key' AND column_name = 'accesed';
	`

	var hasColumn int
	err := tx.QueryRowContext(ctx, hasColumnQuery).Scan(&hasColumn)

	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("failed to check if column accesed exists: %w", err)
		}
	}

	if hasColumn == 0 {
		return nil
	}

	alterTable, err := tx.PrepareContext(ctx, "ALTER TABLE entry_key RENAME COLUMN accesed TO accessed;")

	if err != nil {
		return err
	}

	_, err = alterTable.Exec()

	if err != nil {
		return fmt.Errorf("failed to rename accesed column: %w", err)
	}

	return nil
}

func (e *EntryKeyMigration) Alter(ctx context.Context, tx *sql.Tx) error {
	if err := e.renameAccesedToAccessed(ctx, tx); err != nil {
		return err
	}
	return nil
}
