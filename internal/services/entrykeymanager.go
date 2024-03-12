package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/models"
)

type EntryKeyModel interface {
	Create(ctx context.Context, tx *sql.Tx, entryUUID string, encryptedKey []byte, hash []byte) (*models.EntryKey, error)
	Get(ctx context.Context, db *sql.DB, entryUUID string) ([]models.EntryKey, error)
	Delete(ctx context.Context, tx *sql.Tx, uuid string) error
	SetExpire(ctx context.Context, tx *sql.Tx, uuid string, expire time.Time) error
	SetMaxRead(ctx context.Context, tx *sql.Tx, uuid string, maxRead int) error
	Use(ctx context.Context, tx *sql.Tx, uuid string) error
}

type Hasher interface {
	Hash(data []byte) []byte
}

type EntryKeyManager struct {
	db        *sql.DB
	model     EntryKeyModel
	hasher    Hasher
	encrypter EncrypterFactory
}

func NewEntryKeyManager(db *sql.DB, model EntryKeyModel, hasher Hasher, encrypter EncrypterFactory) *EntryKeyManager {
	return &EntryKeyManager{
		db:        db,
		model:     model,
		hasher:    hasher,
		encrypter: encrypter,
	}
}

func (e *EntryKeyManager) Create(ctx context.Context, entryUUID string, dek []byte, expire *time.Time, maxRead *int) (*models.EntryKey, *key.Key, error) {

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	k, err := key.NewGeneratedKey()

	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	encrypter := e.encrypter(k.Get())
	encryptedKey, err := encrypter.Encrypt(dek)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	hash := e.hasher.Hash(dek)
	entryKey, err := e.model.Create(ctx, tx, entryUUID, encryptedKey, hash)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if expire != nil {
		err := e.model.SetExpire(ctx, tx, entryKey.UUID, *expire)
		if err != nil {
			tx.Rollback()
			return nil, nil, err
		}
		entryKey.Expire = sql.NullTime{
			Time:  *expire,
			Valid: true,
		}
	}

	if maxRead != nil {
		err := e.model.SetMaxRead(ctx, tx, entryKey.UUID, *maxRead)
		if err != nil {
			tx.Rollback()
			return nil, nil, err
		}

		entryKey.RemainingReads = sql.NullInt16{
			Int16: int16(*maxRead),
			Valid: true,
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return entryKey, k, nil
}
