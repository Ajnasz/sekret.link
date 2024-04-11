package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type Encrypter interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

// AESEncrypter is a simple encrypter that uses aes to encrypt and decrypt data
type AESEncrypter struct {
	Key []byte
}

// NewAESEncrypter creates a new Encrypter
func NewAESEncrypter(key []byte) *AESEncrypter {
	return &AESEncrypter{key}
}

func (e *AESEncrypter) WithKey(key []byte) *AESEncrypter {
	e.Key = key
	return e
}

// Encrypt will encrypt the data with the AESEncrypter.Key
func (e *AESEncrypter) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aesGCM.Seal(nonce, nonce, data, nil), nil
}

// Decrypt will dencrypt the data with the AESEncrypter.Key
func (e *AESEncrypter) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
