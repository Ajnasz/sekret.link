package secret

import (
	"context"
	"time"

	"github.com/Ajnasz/sekret.link/encrypter"
	"github.com/Ajnasz/sekret.link/entries"
	"github.com/Ajnasz/sekret.link/storage"
)

// SecretStorage connects the encrypter.Encrypter with a VerifyStorage
// so the encrypted data will be stored in the storage
type SecretStorage struct {
	internalStorage storage.Verifyable
	Encrypter       encrypter.Encrypter
}

// NewSecretStorage creates a secretStore instance
func NewSecretStorage(v storage.Verifyable, e encrypter.Encrypter) *SecretStorage {
	return &SecretStorage{v, e}
}

// Create stores the encrypted secret in the VerifyStorage
func (s SecretStorage) Create(ctx context.Context, UUID string, entry []byte, expire time.Duration, remainingReads int) error {
	encrypted, err := s.Encrypter.Encrypt(entry)

	if err != nil {
		return err
	}

	return s.internalStorage.Create(ctx, UUID, encrypted, expire, remainingReads)
}

// GetMeta returns the entry's metadata
func (s SecretStorage) GetMeta(ctx context.Context, UUID string) (*entries.EntryMeta, error) {
	entryMeta, err := s.internalStorage.GetMeta(ctx, UUID)

	if err != nil {
		return nil, err
	}

	if entryMeta.IsExpired() {
		return nil, entries.ErrEntryExpired
	}

	return entryMeta, nil
}

// GetAndDelete deletes the secret from VerifyStorage
func (s SecretStorage) GetAndDelete(ctx context.Context, UUID string) (*entries.Entry, error) {
	entry, err := s.internalStorage.GetAndDelete(ctx, UUID)

	if err != nil {
		return nil, err
	}

	if entry.IsExpired() {
		return nil, entries.ErrEntryExpired
	}

	if len(entry.Data) == 0 {
		return entry, nil
	}

	decrypted, err := s.Encrypter.Decrypt(entry.Data)

	if err != nil {
		return nil, err
	}

	ret := *entry
	ret.Data = decrypted

	return &ret, nil
}

// VerifyDelete checks if the given deleteKey belongs to the given UUID
func (s SecretStorage) VerifyDelete(ctx context.Context, UUID string, deleteKey string) (bool, error) {
	return s.internalStorage.VerifyDelete(ctx, UUID, deleteKey)
}

// Close Closes the storage connection
func (s SecretStorage) Close() error {
	return s.internalStorage.Close()
}

// Delete Deletes the entry from the storage
func (s SecretStorage) Delete(ctx context.Context, UUID string) error {
	return s.internalStorage.Delete(ctx, UUID)
}

// DeleteExpired removes all expired entries from the storage
func (s SecretStorage) DeleteExpired(ctx context.Context) error {
	return s.internalStorage.DeleteExpired(ctx)
}

// NewCleanableSecretStorage Creates a cleanable secret storage
func NewCleanableSecretStorage(s *SecretStorage, internal storage.Cleanable) CleanableSecretStorage {
	return CleanableSecretStorage{s, internal}
}

// CleanableSecretStorage Storage which implements CleanableStorage interface,
// to allow to clean every entry from the underlying storage
type CleanableSecretStorage struct {
	*SecretStorage
	internalStorage storage.Cleanable
}

// Clean Executes the clean call on the storage
func (s CleanableSecretStorage) Clean() {
	s.internalStorage.Clean()
}
