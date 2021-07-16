package storage

import (
	"time"

	"github.com/Ajnasz/sekret.link/encrypter"
	"github.com/Ajnasz/sekret.link/entries"
)

type VerifyStorage interface {
	EntryStorage
	VerifyDelete(string, string) (bool, error)
}

type SecretStorage struct {
	internalStorage VerifyStorage
	Encrypter       encrypter.Encrypter
}

func NewSecretStorage(v VerifyStorage, e encrypter.Encrypter) *SecretStorage {
	return &SecretStorage{v, e}
}

func (s SecretStorage) Create(UUID string, entry []byte, expire time.Duration, remainingReads int) error {
	encrypted, err := s.Encrypter.Encrypt(entry)

	if err != nil {
		return err
	}

	return s.internalStorage.Create(UUID, encrypted, expire, remainingReads)
}

func (s SecretStorage) GetMeta(UUID string) (*entries.EntryMeta, error) {
	entryMeta, err := s.internalStorage.GetMeta(UUID)

	if err != nil {
		return nil, err
	}

	if entryMeta.IsExpired() {
		return nil, entries.ErrEntryExpired
	}

	return entryMeta, nil
}

func (s SecretStorage) GetAndDelete(UUID string) (*entries.Entry, error) {
	entry, err := s.internalStorage.GetAndDelete(UUID)

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

func (s SecretStorage) VerifyDelete(UUID string, deleteKey string) (bool, error) {
	return s.internalStorage.VerifyDelete(UUID, deleteKey)
}

// Close Closes the storage connection
func (s SecretStorage) Close() error {
	return s.internalStorage.Close()
}

func (s SecretStorage) Delete(UUID string) error {
	return s.internalStorage.Delete(UUID)
}
func (s SecretStorage) DeleteExpired() error {
	return s.internalStorage.DeleteExpired()
}

// CleanableSecretStorage Storage which implements CleanableStorage interface,
// to allow to clean every entry from the underlying storage
type CleanableSecretStorage struct {
	*SecretStorage
	internalStorage CleanableStorage
}

// Clean Executes the clean call on the storage
func (s CleanableSecretStorage) Clean() {
	s.internalStorage.Clean()
}
