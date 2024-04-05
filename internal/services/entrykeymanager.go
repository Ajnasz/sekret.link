package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Ajnasz/sekret.link/internal/hasher"
	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/models"
)

var ErrEntryKeyNotFound = errors.New("entry key not found")

type EntryKeyModel interface {
	Create(ctx context.Context, tx *sql.Tx, entryUUID string, encryptedKey []byte, hash []byte) (*models.EntryKey, error)
	Get(ctx context.Context, tx *sql.Tx, entryUUID string) ([]models.EntryKey, error)
	Delete(ctx context.Context, tx *sql.Tx, uuid string) error
	SetExpire(ctx context.Context, tx *sql.Tx, uuid string, expire time.Time) error
	SetMaxReads(ctx context.Context, tx *sql.Tx, uuid string, maxRead int) error
	Use(ctx context.Context, tx *sql.Tx, uuid string) error
}

type EntryKeyManager struct {
	db        *sql.DB
	tx        *sql.Tx
	model     EntryKeyModel
	hasher    hasher.Hasher
	encrypter EncrypterFactory
}

func NewEntryKeyManager(db *sql.DB, model EntryKeyModel, hasher hasher.Hasher, encrypter EncrypterFactory) *EntryKeyManager {
	return &EntryKeyManager{
		db:        db,
		model:     model,
		hasher:    hasher,
		encrypter: encrypter,
	}
}

func (e *EntryKeyManager) Create(ctx context.Context, entryUUID string, dek []byte, expire *time.Time, maxRead *int) (*EntryKey, *key.Key, error) {

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	entryKey, k, err := e.CreateWithTx(ctx, tx, entryUUID, dek, expire, maxRead)

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return entryKey, k, nil
}

type EntryKey struct {
	UUID           string
	EntryUUID      string
	EncryptedKey   []byte
	KeyHash        []byte
	Created        time.Time
	Expire         time.Time
	RemainingReads int
}

func modelEntryKeyToEntryKey(m *models.EntryKey) *EntryKey {
	return &EntryKey{
		UUID:           m.UUID,
		EntryUUID:      m.EntryUUID,
		EncryptedKey:   m.EncryptedKey,
		KeyHash:        m.KeyHash,
		Created:        m.Created,
		Expire:         m.Expire.Time,
		RemainingReads: int(m.RemainingReads.Int16),
	}
}

func (e *EntryKeyManager) CreateWithTx(ctx context.Context, tx *sql.Tx, entryUUID string, dek []byte, expire *time.Time, maxRead *int) (*EntryKey, *key.Key, error) {
	k, err := key.NewGeneratedKey()

	if err != nil {
		return nil, nil, err
	}
	encrypter := e.encrypter(k.Get())
	encryptedKey, err := encrypter.Encrypt(dek)
	if err != nil {
		return nil, nil, err
	}

	hash := e.hasher.Hash(dek)
	entryKey, err := e.model.Create(ctx, tx, entryUUID, encryptedKey, hash)
	if err != nil {
		return nil, nil, err
	}

	if expire != nil {
		err := e.model.SetExpire(ctx, tx, entryKey.UUID, *expire)
		if err != nil {
			return nil, nil, err
		}
		entryKey.Expire = sql.NullTime{
			Time:  *expire,
			Valid: true,
		}
	}

	if maxRead != nil {
		err := e.model.SetMaxReads(ctx, tx, entryKey.UUID, *maxRead)
		if err != nil {
			return nil, nil, err
		}

		entryKey.RemainingReads = sql.NullInt16{
			Int16: int16(*maxRead),
			Valid: true,
		}
	}
	return modelEntryKeyToEntryKey(entryKey), k, nil
}

func (e *EntryKeyManager) Delete(ctx context.Context, uuid string) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := e.model.Delete(ctx, tx, uuid); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (e *EntryKeyManager) UseTx(ctx context.Context, tx *sql.Tx, entryUUID string) error {
	return e.model.Use(ctx, tx, entryUUID)
}

func (e *EntryKeyManager) findDEK(ctx context.Context, tx *sql.Tx, entryUUID string, key []byte) (dek []byte, entryKey *models.EntryKey, err error) {
	entryKeys, err := e.model.Get(ctx, tx, entryUUID)
	if err != nil {
		return nil, nil, err
	}

	crypter := e.encrypter(key)
	for _, ek := range entryKeys {
		decrypted, err := crypter.Decrypt(ek.EncryptedKey)
		if err != nil {
			continue
		}

		hash := e.hasher.Hash(decrypted)

		if hasher.Compare(hash, ek.KeyHash) {
			return decrypted, &ek, nil
		}
	}

	return nil, nil, ErrEntryKeyNotFound
}

// GetDEK returns the decrypted data encryption key and the entry key
// if the key is not found it returns ErrEntryKeyNotFound
// if the key is found but the hash does not match it returns an error
func (e *EntryKeyManager) GetDEK(ctx context.Context, entryUUID string, key []byte) (dek []byte, entryKey *EntryKey, err error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	dek, entryKey, err = e.GetDEKTx(ctx, tx, entryUUID, key)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return dek, entryKey, nil
}

// GetDEKTx returns the decrypted data encryption key and the entry key
// if the key is not found it returns ErrEntryKeyNotFound
// if the key is found but the hash does not match it returns an error
func (e *EntryKeyManager) GetDEKTx(ctx context.Context, tx *sql.Tx, entryUUID string, key []byte) (dek []byte, entryKey *EntryKey, err error) {
	dek, entryKeyModel, err := e.findDEK(ctx, tx, entryUUID, key)

	if err != nil {
		return nil, nil, err
	}

	if err := validateEntryKey(entryKeyModel); err != nil {
		return nil, nil, err
	}

	if e.model == nil {
		return nil, nil, errors.New("model is nil")
	}

	if err := e.model.Use(ctx, tx, entryKeyModel.UUID); err != nil {
		return nil, nil, err
	}

	return dek, modelEntryKeyToEntryKey(entryKeyModel), nil

}

// GenerateEncryptionKey creates a new key for the entry
func (e EntryKeyManager) GenerateEncryptionKey(ctx context.Context, entryUUID string, existingKey []byte, expire *time.Time, maxRead *int) (*EntryKey, *key.Key, error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	dek, _, err := e.findDEK(ctx, tx, entryUUID, existingKey)

	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	entryKey, k, err := e.CreateWithTx(ctx, tx, entryUUID, dek, expire, maxRead)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return entryKey, k, nil
}
