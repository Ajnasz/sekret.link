package hasher

import (
	"crypto/sha256"
	"crypto/subtle"
)

type Hasher interface {
	Hash(data []byte) []byte
}

type Sha256Hasher struct{}

func NewSHA256Hasher() *Sha256Hasher {
	return &Sha256Hasher{}
}

func (h *Sha256Hasher) Hash(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

func Compare(k, k2 []byte) bool {
	return subtle.ConstantTimeCompare(k, k2) == 1
}
