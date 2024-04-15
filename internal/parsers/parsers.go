package parsers

import (
	"net/http"

	"github.com/Ajnasz/sekret.link/internal/key"
)

type Parser[T any] interface {
	Parse(r *http.Request) (T, error)
}

func getEntryKeyByte(keyString string) ([]byte, error) {
	if len(keyString) == 64 {
		return key.FromHex(keyString)
	}
	return nil, ErrInvalidKeyLength
}
