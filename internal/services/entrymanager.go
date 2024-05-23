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

var ErrCreateEntryFailed = errors.New("create entry failed")
var ErrReadEntryFailed = errors.New("read entry failed")
var ErrDeleteEntryFailed = errors.New("delete entry failed")
var DeleteExpiredFailed = errors.New("delete expired failed")

// EntryMeta provides the entry meta
type EntryMeta struct {
	UUID           string
	RemainingReads int
	DeleteKey      string
	Created        time.Time
	Accessed       time.Time
	Expire         time.Time
	ContentType    string
}

type Entry struct {
	EntryMeta
	Data []byte
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
func (e *EntryManager) CreateEntry(ctx context.Context, contentType string, data []byte, remainingReads int, expire time.Duration) (*EntryMeta, key.Key, error) {
	uid := uuid.NewUUIDString()

	tx, err := e.db.Begin()
	if err != nil {
		return nil, nil, errors.Join(ErrCreateEntryFailed, err)
	}
	dek, err := key.NewGeneratedKey()

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, nil, errors.Join(ErrCreateEntryFailed, err, rollbackErr)
		}
		return nil, nil, errors.Join(ErrCreateEntryFailed, err)
	}

	crypto := e.crypto(dek.Get())

	encryptedData, err := crypto.Encrypt(data)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, nil, errors.Join(ErrCreateEntryFailed, err, rollbackErr)
		}
		return nil, nil, errors.Join(ErrCreateEntryFailed, err)
	}
	meta, err := e.model.CreateEntry(ctx, tx, uid, contentType, encryptedData)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, nil, errors.Join(ErrCreateEntryFailed, err, rollbackErr)
		}
		return nil, nil, errors.Join(ErrCreateEntryFailed, err)
	}

	expireAt := time.Now().Add(expire)
	entryKey, kek, err := e.keyManager.CreateWithTx(ctx, tx, uid, dek.Get(), expireAt, remainingReads)

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, nil, errors.Join(ErrCreateEntryFailed, err, rollbackErr)
		}
		return nil, nil, errors.Join(ErrCreateEntryFailed, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, errors.Join(ErrCreateEntryFailed, err)
	}

	return &EntryMeta{
		UUID:           meta.UUID,
		DeleteKey:      meta.DeleteKey,
		Created:        meta.Created,
		Accessed:       meta.Accessed.Time,
		ContentType:    meta.ContentType,
		RemainingReads: entryKey.RemainingReads,
		Expire:         entryKey.Expire,
	}, kek, nil
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
		if err := tx.Rollback(); err != nil {
			return nil, errors.Join(err, ErrReadEntryFailed)
		}
		if errors.Is(err, models.ErrEntryNotFound) {
			return nil, ErrEntryNotFound
		}
		return nil, errors.Join(err, ErrReadEntryFailed)
	}

	if err := validateEntry(entry); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, errors.Join(err, ErrReadEntryFailed)
		}
		return nil, err
	}

	dek, entryKey, err := e.keyManager.GetDEKTx(ctx, tx, UUID, k)
	var decryptedData []byte
	if err != nil {
		if errors.Is(err, ErrEntryKeyNotFound) {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, errors.Join(err, rollbackErr, ErrReadEntryFailed)
			}
			return nil, errors.Join(err, ErrReadEntryFailed)
		} else {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, errors.Join(err, rollbackErr, ErrReadEntryFailed)
			}
			return nil, errors.Join(err, ErrReadEntryFailed)
		}
	} else {
		crypto := e.crypto(dek)
		decryptedData, err = crypto.Decrypt(entry.Data)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, errors.Join(err, rollbackErr, ErrReadEntryFailed)
			}
			return nil, errors.Join(err, ErrReadEntryFailed)
		}

		if err := e.keyManager.UseTx(ctx, tx, entryKey.UUID); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, errors.Join(err, rollbackErr, ErrReadEntryFailed)
			}
			return nil, errors.Join(err, ErrReadEntryFailed)
		}
	}

	if err := e.model.Use(ctx, tx, UUID); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, errors.Join(err, rollbackErr, ErrReadEntryFailed)
		}
		return nil, errors.Join(err, ErrReadEntryFailed)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Join(err, ErrReadEntryFailed)
	}

	return &Entry{
		EntryMeta: EntryMeta{
			UUID:           entry.UUID,
			DeleteKey:      entry.DeleteKey,
			Created:        entry.Created,
			Accessed:       entry.Accessed.Time,
			ContentType:    entry.ContentType,
			Expire:         entryKey.Expire,
			RemainingReads: entryKey.RemainingReads,
		},
		Data: decryptedData,
	}, nil
}

func (e *EntryManager) DeleteEntry(ctx context.Context, UUID string, deleteKey string) error {
	tx, err := e.db.Begin()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(ErrDeleteEntryFailed, err, rollbackErr)
		}
		return errors.Join(ErrDeleteEntryFailed, err)
	}

	if err := e.model.DeleteEntry(ctx, tx, UUID, deleteKey); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(ErrDeleteEntryFailed, err, rollbackErr)
		}
		return errors.Join(ErrDeleteEntryFailed, err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Join(ErrDeleteEntryFailed, err)
	}
	return nil
}

func (e *EntryManager) DeleteExpired(ctx context.Context) error {
	tx, err := e.db.Begin()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(DeleteExpiredFailed, err, rollbackErr)
		}
		return errors.Join(DeleteExpiredFailed, err)
	}

	if err := e.model.DeleteExpired(ctx, tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(DeleteExpiredFailed, err, rollbackErr)
		}
		return errors.Join(DeleteExpiredFailed, err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Join(DeleteExpiredFailed, err)
	}
	return nil
}

func (e *EntryManager) GenerateEntryKey(ctx context.Context, entryUUID string, k key.Key, expire time.Duration, maxReads int) (*EntryKeyData, error) {
	meta, kek, err := e.keyManager.GenerateEncryptionKey(ctx, entryUUID, k, time.Now().Add(expire), maxReads)
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
