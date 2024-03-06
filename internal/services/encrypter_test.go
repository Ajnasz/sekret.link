package services

import (
	"testing"
)

func Test_Encrypter_Encrypt(t *testing.T) {
	testData := "Lorem ipsum dolor sit amet"
	encKey := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	encrypter := NewAESEncrypter(encKey)

	data, err := encrypter.Encrypt([]byte(testData))

	if err != nil {
		t.Fatal(err)
	}

	if data == nil {
		t.Error("Encrypted data is nil")
	}

	if len(data) == 0 {
		t.Error("Encrypted data is empty")
	}

	if string(data) == testData {
		t.Error("Encrypted data is not encrypted")
	}
}

func Test_Encrypter_Decrypt(t *testing.T) {
	testData := "Lorem ipsum dolor sit amet"
	encKey := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	encrypter := NewAESEncrypter(encKey)

	data, err := encrypter.Encrypt([]byte(testData))

	if err != nil {
		t.Fatal(err)
	}

	if data == nil {
		t.Error("Encrypted data is nil")
	}

	decrypted, err := encrypter.Decrypt(data)

	if err != nil {
		t.Fatal(err)
	}

	actual := string(decrypted)
	if string(decrypted) != testData {
		t.Errorf("Decryption failed, expected %q, got %s", testData, actual)
	}
}
