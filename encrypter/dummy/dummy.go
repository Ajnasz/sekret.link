package dummy

import "encoding/base64"

// Encrypter encrypter implementation for tests
type Encrypter struct{}

// Encrypt encrypts data
func (d *Encrypter) Encrypt(data []byte) ([]byte, error) {
	output := make([]byte, base64.RawStdEncoding.EncodedLen(len(data)))
	base64.RawStdEncoding.Encode(output, data)

	return output, nil
}

// Decrypt denrypts data
func (d *Encrypter) Decrypt(data []byte) ([]byte, error) {
	output := make([]byte, base64.RawStdEncoding.DecodedLen(len(data)))
	_, err := base64.RawStdEncoding.Decode(output, data)

	if err != nil {
		return nil, err
	}
	return output, nil
}

// NewEncrypter creates a new DummyEncrypter instance
func NewEncrypter() *Encrypter {
	return &Encrypter{}
}
