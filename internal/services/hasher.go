package services

import "crypto/sha256"

type Sha256Hasher struct{}

func NewSha256Hasher() *Sha256Hasher {
	return &Sha256Hasher{}
}

func (h *Sha256Hasher) Hash(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}
