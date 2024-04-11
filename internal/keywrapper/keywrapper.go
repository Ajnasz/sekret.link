package services

import (
	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/services"
)

// KeyWrapper is a simple interface to wrap and unwrap keys
// rfc3394
type KeyWrapper interface {
	Wrap(dek *key.Key) ([]byte, *key.Key, error)
	Unwrap(kek *key.Key, encrypted []byte) (*key.Key, error)
}

// AesKeyWrapper is a simple key wrapper that uses AES to wrap and unwrap keys
type AesKeyWrapper struct{}

// NewAesKeyWrapper creates a new AesKeyWrapper
func NewAesKeyWrapper() *AesKeyWrapper {
	return &AesKeyWrapper{}
}

// Wrap will wrap the Data Encryption Key (DEK) with the Key Encryption Key (KEK)
func (*AesKeyWrapper) Wrap(dek *key.Key) ([]byte, *key.Key, error) {
	kek := key.NewKey()
	err := kek.Generate()
	if err != nil {
		return nil, nil, err
	}

	encrypter := services.NewAESEncrypter(kek.Get())

	encrypted, err := encrypter.Encrypt(dek.Get())
	if err != nil {
		return nil, nil, err
	}

	return encrypted, kek, nil
}

// Unwrap will unwrap the Data Encryption Key (DEK) with the Key Encryption Key (KEK)
func (*AesKeyWrapper) Unwrap(kek *key.Key, encrypted []byte) (*key.Key, error) {
	decrypter := services.NewAESEncrypter(kek.Get())

	decrypted, err := decrypter.Decrypt(encrypted)
	if err != nil {
		return nil, err
	}

	k := key.NewKey()
	if err := k.Set(decrypted); err != nil {
		return nil, err
	}

	return k, nil
}
