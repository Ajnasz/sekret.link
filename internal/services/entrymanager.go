package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/uuid"
)

var ErrEntryExpired = errors.New("entry expired")
var ErrEntryNotFound = errors.New("entry not found")
var ErrEntryNoRemainingReads = errors.New("entry has no remaining reads")

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

type EntryKeyData struct {
	EntryUUID      string
	KEK            key.Key
	RemainingReads int
	Expire         time.Time
}

// EntryManager provides the entry service
type EntryManager struct {
	db         *sql.DB
	model      EntryModel
	crypto     EncrypterFactory
	keyManager EntryKeyer
}

// NewEntryManager creates a new EntryService
func NewEntryManager(db *sql.DB, model EntryModel, crypto EncrypterFactory, keyManager EntryKeyer) *EntryManager {
	return &EntryManager{
		db:         db,
		model:      model,
		crypto:     crypto,
		keyManager: keyManager,
	}
}

// CreateEntry creates a new entry
// It generates a new UUID for the entry
// It encrypts the data with a new generated key
// It stores the encrypted data in the database
// It stores the key in the key manager
// It returns the meta data of the entry and the key
func (e *EntryManager) CreateEntry(ctx context.Context, data []byte, remainingReads int, expire time.Duration) (*EntryMeta, key.Key, error) {
	uid := uuid.NewUUIDString()

	tx, err := e.db.Begin()
	if err != nil {
		return nil, nil, err
	}
	dek, err := key.NewGeneratedKey()

	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	crypto := e.crypto(dek.Get())

	encryptedData, err := crypto.Encrypt(data)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	meta, err := e.model.CreateEntry(ctx, tx, uid, encryptedData, remainingReads, expire)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	expireAt := time.Now().Add(expire)
	_, kek, err := e.keyManager.CreateWithTx(ctx, tx, uid, dek.Get(), &expireAt, &remainingReads)

	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tx.Commit()

	return &EntryMeta{
		UUID:           meta.UUID,
		RemainingReads: meta.RemainingReads,
		DeleteKey:      meta.DeleteKey,
		Created:        meta.Created,
		Accessed:       meta.Accessed.Time,
		Expire:         meta.Expire,
	}, kek, nil
}

func (e *EntryManager) readEntryLegacy(ctx context.Context, k key.Key, entry *models.Entry) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	crypto := e.crypto(k)
	return crypto.Decrypt(entry.Data)
}

// ReadEntry reads an entry
// It reads the entry from the database
// It reads the key from the key manager
// It decrypts the data with the key
// It returns the decrypted data
// It returns an error if the entry is not found or expired
// It returns an error if the key is not found
func (e *EntryManager) ReadEntry(ctx context.Context, UUID string, k key.Key) (*Entry, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
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

	if err := validateEntry(entry); err != nil {
		tx.Rollback()
		return nil, err
	}

	dek, entryKey, err := e.keyManager.GetDEKTx(ctx, tx, UUID, k)
	var decryptedData []byte
	if err != nil {
		if errors.Is(err, ErrEntryKeyNotFound) {
			legacyData, legacyErr := e.readEntryLegacy(ctx, k, entry)
			if legacyErr == nil {
				decryptedData = legacyData
			} else {
				tx.Rollback()
				return nil, err
			}
		} else {
			tx.Rollback()
			return nil, err
		}
	} else {
		crypto := e.crypto(dek)
		decryptedData, err = crypto.Decrypt(entry.Data)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		if err := e.keyManager.UseTx(ctx, tx, entryKey.UUID); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := e.model.Use(ctx, tx, UUID); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

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

func (e *EntryManager) DeleteExpired(ctx context.Context) error {
	tx, err := e.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := e.model.DeleteExpired(ctx, tx); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (e *EntryManager) GenerateEntryKey(ctx context.Context, entryUUID string, k key.Key) (*EntryKeyData, error) {
	meta, kek, err := e.keyManager.GenerateEncryptionKey(ctx, entryUUID, k, nil, nil)
	if err != nil {
		return nil, err
	}

	return &EntryKeyData{
		EntryUUID:      entryUUID,
		RemainingReads: meta.RemainingReads,
		Expire:         meta.Expire,
		KEK:            kek,
	}, nil
}
