package parsers

import (
	"net/http"

	"github.com/Ajnasz/sekret.link/internal/key"
)

type Parser[T any] interface {
	Parse(r *http.Request) (T, error)
}

func getEntryKeyByte(keyString string) (*key.Key, error) {
	if len(keyString) == 64 {
		return key.FromHex(keyString)
	} else if len(keyString) == 43 {
		return key.FromBase62(keyString)
	}
	return nil, ErrInvalidKeyLength
}
