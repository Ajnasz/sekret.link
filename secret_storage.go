package main

type SecretStorage struct {
	internalStorage EntryStorage
	Encrypter       Encrypter
}

func (s *SecretStorage) Create(UUID string, entry []byte) error {
	encrypted, err := s.Encrypter.Encrypt(entry)

	if err != nil {
		return err
	}

	return s.internalStorage.Create(UUID, encrypted)
}

func (s *SecretStorage) GetMeta(UUID string) (*EntryMeta, error) {
	return s.internalStorage.GetMeta(UUID)
}

func (s *SecretStorage) Get(UUID string) (*Entry, error) {
	entry, err := s.internalStorage.Get(UUID)

	if err != nil {
		return nil, err
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
