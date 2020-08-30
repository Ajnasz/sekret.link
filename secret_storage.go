package main

import (
	"time"
)

type SecretStorage struct {
	internalStorage EntryStorage
	Encrypter       Encrypter
}

func (s *SecretStorage) Create(UUID string, entry []byte, expire time.Duration) error {
	encrypted, err := s.Encrypter.Encrypt(entry)

	if err != nil {
		return err
	}

	return s.internalStorage.Create(UUID, encrypted, expire)
}

func (s *SecretStorage) GetMeta(UUID string) (*EntryMeta, error) {
	entryMeta, err := s.internalStorage.GetMeta(UUID)

	if err != nil {
		return nil, err
	}

	if entryMeta.IsExpired() {
		return nil, ErrEntryExpired
	}

	return entryMeta, nil
}

func (s *SecretStorage) Get(UUID string) (*Entry, error) {
	entry, err := s.internalStorage.Get(UUID)

	if err != nil {
		return nil, err
	}

	if entry.IsExpired() {
		return nil, ErrEntryExpired
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

func (s *SecretStorage) GetAndDelete(UUID string) (*Entry, error) {
	entry, err := s.internalStorage.GetAndDelete(UUID)

	if err != nil {
		return nil, err
	}

	if entry.IsExpired() {
		return nil, ErrEntryExpired
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

func (s *SecretStorage) Close() error {
	return s.internalStorage.Close()
}

func (s *SecretStorage) Delete(UUID string) error {
	return s.internalStorage.Delete(UUID)
}
func (s *SecretStorage) DeleteExpired() error {
	return s.internalStorage.DeleteExpired()
}
