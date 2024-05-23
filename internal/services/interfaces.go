package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/models"
)

// EntryModel is the interface for the entry model
// It is used to create, read and access entries
type EntryModel interface {
	CreateEntry(ctx context.Context, tx *sql.Tx, UUID string, contentType string, data []byte) (*models.EntryMeta, error)
	ReadEntry(ctx context.Context, tx *sql.Tx, UUID string) (*models.Entry, error)
	Use(ctx context.Context, tx *sql.Tx, UUID string) error
	DeleteEntry(ctx context.Context, tx *sql.Tx, UUID string, deleteKey string) error
	DeleteExpired(ctx context.Context, tx *sql.Tx) error
}

// EntryKeyer is the interface for the entry key manager
// It is used to create, read and access entry keys
type EntryKeyer interface {
	CreateWithTx(ctx context.Context, tx *sql.Tx, entryUUID string, dek key.Key, expire time.Time, maxRead int) (entryKey *EntryKey, kek key.Key, err error)
	GetDEKTx(ctx context.Context, tx *sql.Tx, entryUUID string, kek key.Key) (dek key.Key, entryKey *EntryKey, err error)
	GenerateEncryptionKey(ctx context.Context, entryUUID string, existingKey key.Key, expire time.Time, maxRead int) (*EntryKey, key.Key, error)
	UseTx(ctx context.Context, tx *sql.Tx, entryUUID string) error
}

// EncrypterFactory is function to create a new Encrypter for a given key
type EncrypterFactory = func(k key.Key) Encrypter
