package hasher

import "crypto/sha256"

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
	if len(k) != len(k2) {
		return false
	}
	for i := range k {
		if k[i] != k2[i] {
			return false
		}
	}
	return true
}
