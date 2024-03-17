package parsers

import (
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers/expiration"
	"github.com/google/uuid"
)

type GenerateEntryKeyRequestData struct {
	UUID      string
	KeyString string
	Key       []byte
}

type GenerateEntryKeyParser struct {
	maxExpireSeconds int
	getEntryParser   GetEntryParser
}

func NewGenerateEntryKeyParser() *GenerateEntryKeyParser {
	return &GenerateEntryKeyParser{
		getEntryParser: NewGetEntryParser(),
	}
}

func (g GenerateEntryKeyParser) calculateExpiration(expire string, defaultExpire time.Duration) (time.Duration, error) {
	exp, err := expiration.CalculateExpiration(expire, defaultExpire, g.maxExpireSeconds)
	if err != nil {
		return 0, ErrInvalidExpirationDate
	}

	return exp, nil
}

func (g GenerateEntryKeyParser) getSecretExpiration(req *http.Request) (time.Duration, error) {
	expiration := req.URL.Query().Get("expiration")
	if expiration == "" {
		return 0, nil
	}
	return 0, nil
}

func (p *GenerateEntryKeyParser) Parse(r *http.Request) (GenerateEntryKeyRequestData, error) {
	var reqData GenerateEntryKeyRequestData
	keyString := r.PathValue("key")
	if keyString == "" {
		return reqData, ErrInvalidKey
	}

	uuidFromPath := r.PathValue("uuid")
	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return reqData, errors.Join(ErrInvalidUUID, err)
	}

	reqData.UUID = UUID.String()
	reqData.KeyString = keyString
	reqData.Key, err = hex.DecodeString(keyString)

	return reqData, nil
}
