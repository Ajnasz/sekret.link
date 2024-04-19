package services

import (
	"context"
	"database/sql"
)

type ExpiredEntryModel interface {
	DeleteExpired(ctx context.Context, tx *sql.Tx) error
}

type ExpiredEntryManager struct {
	db            *sql.DB
	entryModel    ExpiredEntryModel
	entryKeyModel ExpiredEntryModel
}

func NewExpiredEntryManager(db *sql.DB, entryModel ExpiredEntryModel, entryKeyModel ExpiredEntryModel) *ExpiredEntryManager {
	return &ExpiredEntryManager{
		db:            db,
		entryModel:    entryModel,
		entryKeyModel: entryKeyModel,
	}
}

func (d *ExpiredEntryManager) DeleteExpired(ctx context.Context) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	if err := d.entryKeyModel.DeleteExpired(ctx, tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := d.entryModel.DeleteExpired(ctx, tx); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
