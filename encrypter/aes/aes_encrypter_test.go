package aes

import (
	"testing"
)

func TestAESEncrypter(t *testing.T) {
	testData := "Lorem ipsum dolor sit amet"
	encKey := []byte("B629C0156C0FEBDF1B1DC244052E3EC9")

	encrypter := Encrypter{encKey}

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

func TestAESEncrypterBase64(t *testing.T) {
	testData := "Lorem ipsum dolor sit amet"
	encKey := []byte("B629C0156C0FEBDF1B1DC244052E3EC9")

	encrypter := Encrypter{encKey}

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
