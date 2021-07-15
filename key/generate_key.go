package key

import (
	"crypto/rand"
)

func GenerateRSAKey() ([]byte, error) {
	bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return bytes, nil

}
