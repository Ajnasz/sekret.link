package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/uuid"
)

// EntryModel is the interface for the entry model
// It is used to create, read and access entries
type EntryModel interface {
	CreateEntry(ctx context.Context, tx *sql.Tx, UUID string, data []byte, remainingReads int, expire time.Duration) (*models.EntryMeta, error)
	ReadEntry(ctx context.Context, tx *sql.Tx, UUID string) (*models.Entry, error)
	UpdateAccessed(ctx context.Context, tx *sql.Tx, UUID string) error
}

// EntryCrypto is the interface to encrypt and decrypt the entry data
type EntryCrypto interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

// EntryMeta provides the entry meta
type EntryMeta struct {
	UUID           string
	RemainingReads int
	DeleteKey      string
	Created        time.Time
	Accessed       time.Time
	Expire         time.Time
}

// EntryManager provides the entry service
type EntryManager struct {
	db     *sql.DB
	model  EntryModel
	crypto EntryCrypto
}

// NewEntry creates a new EntryService
func NewEntry(db *sql.DB, model EntryModel, crypto EntryCrypto) *EntryManager {
	return &EntryManager{
		db:     db,
		model:  model,
		crypto: crypto,
	}
}

func (e *EntryManager) CreateEntry(ctx context.Context, data []byte, remainingReads int, expire time.Duration) (*EntryMeta, error) {
	uid := uuid.NewUUIDString()

	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}

	encryptedData, err := e.crypto.Encrypt(data)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// meta, err := e.model.CreateEntry(ctx, tx, uid, data, remainingReads, expire)
	meta, err := e.model.CreateEntry(ctx, tx, uid, encryptedData, remainingReads, expire)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return &EntryMeta{
		UUID:           meta.UUID,
		RemainingReads: meta.RemainingReads,
		DeleteKey:      meta.DeleteKey,
		Created:        meta.Created,
		Accessed:       meta.Accessed.Time,
		Expire:         meta.Expire,
	}, nil

}

func (e *EntryManager) ReadEntry(ctx context.Context, UUID string) ([]byte, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}

	entry, err := e.model.ReadEntry(ctx, tx, UUID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := e.model.UpdateAccessed(ctx, tx, UUID); err != nil {
		tx.Rollback()
		return nil, err
	}

	decryptedData, err := e.crypto.Decrypt(entry.Data)

	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return decryptedData, nil
}
