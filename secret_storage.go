package main

type SecretStorage struct {
	internalStorage EntryStorage
	encrypter       Encrypter
}

func (s *SecretStorage) Create(UUID string, entry []byte) error {
	encrypted, err := s.encrypter.Encrypt(entry)

	if err != nil {
		return err
	}

	return s.internalStorage.Create(UUID, encrypted)
}

func (s *SecretStorage) Get(UUID string) ([]byte, error) {
	data, err := s.internalStorage.Get(UUID)

	if err != nil {
		return nil, err
	}

	return s.encrypter.Decrypt(data)
}

func (s *SecretStorage) GetAndDelete(UUID string) ([]byte, error) {
	data, err := s.internalStorage.GetAndDelete(UUID)

	if err != nil {
		return nil, err
	}

	return s.encrypter.Decrypt(data)
}
