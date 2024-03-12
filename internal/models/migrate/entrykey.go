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
	key_hash CHAR(64) NOT NULL,
	created TIMESTAMPTZ,
	expire TIMESTAMPTZ DEFAULT NULL,
	remaining_reads SMALLINT DEFAULT NULL,
	FOREIGN KEY (entry_uuid) REFERENCES entries(uuid) ON DELETE CASCADE
	);
`
	_, err := tx.ExecContext(ctx, query)

	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (e *EntryKeyMigration) Alter(ctx context.Context, tx *sql.Tx) error {
	return nil
}
