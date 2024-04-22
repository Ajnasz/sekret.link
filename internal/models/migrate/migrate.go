package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

type Migrator interface {
	Create(context.Context, *sql.Tx) error
	Alter(context.Context, *sql.Tx) error
}

func prepareDatabase(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	migrations := []Migrator{
		NewEntryMigration(),
		NewEntryKeyMigration(),
	}

	for _, migration := range migrations {
		if err := migration.Create(ctx, tx); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("failed to create migration: %w", rollbackErr)
			}
			return fmt.Errorf("failed to create migration: %w", err)
		}

		if err := migration.Alter(ctx, tx); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("failed to alter migration: %w", rollbackErr)
			}
			return fmt.Errorf("failed to alter migration: %w", err)
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
