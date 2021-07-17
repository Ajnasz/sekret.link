package storage

import "encoding/base64"

type DummyEncrypter struct{}

func (d *DummyEncrypter) Encrypt(data []byte) ([]byte, error) {
	output := make([]byte, base64.RawStdEncoding.EncodedLen(len(data)))
	base64.RawStdEncoding.Encode(output, data)

	return output, nil
}
func (d *DummyEncrypter) Decrypt(data []byte) ([]byte, error) {
	output := make([]byte, base64.RawStdEncoding.DecodedLen(len(data)))
	_, err := base64.RawStdEncoding.Decode(output, data)

	if err != nil {
		return nil, err
	}
	return output, nil
}

func NewDummyEncrypter() *DummyEncrypter {
	return &DummyEncrypter{}
}
