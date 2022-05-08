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
	internalStorage storage.VerifyConfirmReader
	encrypter       encrypter.Encrypter
}

// NewSecretStorage creates a secretStore instance
func NewSecretStorage(v storage.VerifyConfirmReader, e encrypter.Encrypter) *SecretStorage {
	return &SecretStorage{v, e}
}

// Write stores the encrypted secret in the VerifyStorage
func (s SecretStorage) Write(ctx context.Context, UUID string, entry []byte, expire time.Duration, remainingReads int) (*entries.EntryMeta, error) {
	encrypted, err := s.encrypter.Encrypt(entry)

	if err != nil {
		return nil, err
	}

	return s.internalStorage.Write(ctx, UUID, encrypted, expire, remainingReads)
}

// Read deletes the secret from VerifyStorage
func (s SecretStorage) Read(ctx context.Context, UUID string) (*entries.Entry, error) {
	entry, confirm, err := s.internalStorage.ReadConfirm(ctx, UUID)

	if err != nil {
		return nil, err
	}

	if entry.IsExpired() {
		confirm <- false
		return nil, entries.ErrEntryExpired
	}

	if len(entry.Data) == 0 {
		confirm <- true
		<-confirm
		return entry, nil
	}

	decrypted, err := s.encrypter.Decrypt(entry.Data)

	if err != nil {
		confirm <- false
		return nil, err
	}

	confirm <- true
	select {
	case <-confirm:
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
