package services

import (
	"testing"

	"github.com/Ajnasz/sekret.link/internal/key"
)

func Test_KeyWrap(t *testing.T) {
	k := key.NewKey()

	err := k.Generate()
	if err != nil {
		t.Fatal(err)
	}

	keyWrapper := NewAesKeyWrapper()

	encrypted, kek, err := keyWrapper.Wrap(k)
	if err != nil {
		t.Fatal(err)
	}

	badKek := key.NewKey()
	err = badKek.Generate()
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := keyWrapper.Unwrap(kek, encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if string(decrypted.Get()) != string(k.Get()) {
		t.Errorf("key unwrap failed, expected %s, got %s", k.Get(), decrypted.Get())
	}

	_, err = keyWrapper.Unwrap(badKek, encrypted)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
