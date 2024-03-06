package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/uuid"
)

var ErrEntryExpired = errors.New("entry expired")
var ErrEntryNotFound = errors.New("entry not found")

// EntryModel is the interface for the entry model
// It is used to create, read and access entries
type EntryModel interface {
	CreateEntry(ctx context.Context, tx *sql.Tx, UUID string, data []byte, remainingReads int, expire time.Duration) (*models.EntryMeta, error)
	ReadEntry(ctx context.Context, tx *sql.Tx, UUID string) (*models.Entry, error)
	UpdateAccessed(ctx context.Context, tx *sql.Tx, UUID string) error
	DeleteEntry(ctx context.Context, tx *sql.Tx, UUID string, deleteKey string) error
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

type Entry struct {
	UUID           string
	Data           []byte
	RemainingReads int
	DeleteKey      string
	Created        time.Time
	Accessed       time.Time
	Expire         time.Time
}

func (e *Entry) IsExpired() bool {
	return e.Expire.Before(time.Now())
}

// EntryManager provides the entry service
type EntryManager struct {
	db     *sql.DB
	model  EntryModel
	crypto Encrypter
}

// NewEntryManager creates a new EntryService
func NewEntryManager(db *sql.DB, model EntryModel, crypto Encrypter) *EntryManager {
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

func (e *EntryManager) ReadEntry(ctx context.Context, UUID string) (*Entry, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}

	entry, err := e.model.ReadEntry(ctx, tx, UUID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, models.ErrEntryNotFound) {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	if entry.RemainingReads <= 0 {
		tx.Rollback()
		return nil, ErrEntryExpired
	}

	if entry.Expire.Before(time.Now()) {
		tx.Rollback()
		return nil, ErrEntryExpired
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

	return &Entry{
		UUID:           entry.UUID,
		Data:           decryptedData,
		RemainingReads: entry.RemainingReads - 1,
		DeleteKey:      entry.DeleteKey,
		Created:        entry.Created,
		Accessed:       entry.Accessed.Time,
		Expire:         entry.Expire,
	}, nil
}

func (e *EntryManager) DeleteEntry(ctx context.Context, UUID string, deleteKey string) error {
	tx, err := e.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := e.model.DeleteEntry(ctx, tx, UUID, deleteKey); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
