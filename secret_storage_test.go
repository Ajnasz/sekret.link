package main

import (
	"encoding/base64"
	"testing"
	"time"
)

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

func TestSecretStorage(t *testing.T) {

	testData := "Lorem ipusm dolor sit amet"
	storage := &secretStorage{newMemoryStorage(), NewDummyEncrypter()}

	UUID := newUUIDString()
	err := storage.Create(UUID, []byte(testData), time.Second*10, 1)

	if err != nil {
		t.Fatal(err)
	}

	data, err := storage.Get(UUID)

	if err != nil {
		t.Fatal(err)
	}

	actual := string(data.Data)

	if actual != testData {
		t.Errorf("Expected %q, actual %q", testData, actual)
	}
}
