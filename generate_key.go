package main

import (
	"crypto/rand"
)

func generateRSAKey() ([]byte, error) {
	bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return bytes, nil

}
