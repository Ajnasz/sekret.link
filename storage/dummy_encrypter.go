package storage

import "encoding/base64"

// DummyEncrypter encrypter implementation for tests
type DummyEncrypter struct{}

// Encrypt encrypts data
func (d *DummyEncrypter) Encrypt(data []byte) ([]byte, error) {
	output := make([]byte, base64.RawStdEncoding.EncodedLen(len(data)))
	base64.RawStdEncoding.Encode(output, data)

	return output, nil
}

// Decrypt denrypts data
func (d *DummyEncrypter) Decrypt(data []byte) ([]byte, error) {
	output := make([]byte, base64.RawStdEncoding.DecodedLen(len(data)))
	_, err := base64.RawStdEncoding.Decode(output, data)

	if err != nil {
		return nil, err
	}
	return output, nil
}

// NewDummyEncrypter creates a new DummyEncrypter instance
func NewDummyEncrypter() *DummyEncrypter {
	return &DummyEncrypter{}
}
