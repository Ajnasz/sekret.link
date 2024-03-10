package services

import (
	"context"
	"database/sql"
)

type ExpiredEntryModel interface {
	DeleteExpired(ctx context.Context, tx *sql.Tx) error
}

type ExpiredEntryManager struct {
	db    *sql.DB
	model ExpiredEntryModel
}

func NewExpiredEntryManager(db *sql.DB, model ExpiredEntryModel) *ExpiredEntryManager {
	return &ExpiredEntryManager{
		db:    db,
		model: model,
	}
}

func (d *ExpiredEntryManager) DeleteExpired(ctx context.Context) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	if err := d.model.DeleteExpired(ctx, tx); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
