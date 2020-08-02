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

func (s *SecretStorage) Get(UUID string) ([]byte, error) {
	data, err := s.internalStorage.Get(UUID)

	if len(data) == 0 {
		return data, nil
	}

	if err != nil {
		return nil, err
	}

	return s.Encrypter.Decrypt(data)
}

func (s *SecretStorage) GetAndDelete(UUID string) ([]byte, error) {
	data, err := s.internalStorage.GetAndDelete(UUID)

	if len(data) == 0 {
		return data, nil
	}

	if err != nil {
		return nil, err
	}

	return s.Encrypter.Decrypt(data)
}
