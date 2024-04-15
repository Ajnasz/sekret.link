package parsers

import (
	"errors"
	"net/http"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers/expiration"
	"github.com/Ajnasz/sekret.link/internal/parsers/maxreads"
	"github.com/google/uuid"
)

// GenerateEntryKeyRequestData is the data for the GenerateEntryKey endpoint.
type GenerateEntryKeyRequestData struct {
	UUID       string
	Key        []byte
	Expiration time.Duration
	MaxReads   int
}

// GenerateEntryKeyParser is the http request parser for the GenerateEntryKey endpoint.
type GenerateEntryKeyParser struct {
	maxExpireSeconds int
}

// NewGenerateEntryKeyParser returns a new GenerateEntryKeyParser.
func NewGenerateEntryKeyParser(maxExpireSeconds int) *GenerateEntryKeyParser {
	return &GenerateEntryKeyParser{
		maxExpireSeconds: maxExpireSeconds,
	}
}

func (g GenerateEntryKeyParser) calculateExpiration(expire string, defaultExpire time.Duration) (time.Duration, error) {
	exp, err := expiration.Parse(expire, defaultExpire, g.maxExpireSeconds)
	if err != nil {
		return 0, ErrInvalidExpirationDate
	}

	return exp, nil
}

func (g GenerateEntryKeyParser) getSecretExpiration(req *http.Request) (time.Duration, error) {
	expiration := req.URL.Query().Get("expire")

	return g.calculateExpiration(expiration, time.Second*time.Duration(g.maxExpireSeconds))
}

func (g GenerateEntryKeyParser) getSecretMaxReads(req *http.Request) (int, error) {
	maxReads := req.URL.Query().Get("maxReads")

	return maxreads.Parse(maxReads)
}

// Parse parses the http request for the GenerateEntryKey endpoint.
func (g *GenerateEntryKeyParser) Parse(r *http.Request) (GenerateEntryKeyRequestData, error) {
	var reqData GenerateEntryKeyRequestData
	uuidFromPath := r.PathValue("uuid")
	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return reqData, errors.Join(ErrInvalidUUID, err)
	}

	keyString := r.PathValue("key")
	if keyString == "" {
		return reqData, ErrInvalidKey
	}

	keyByte, err := getEntryKeyByte(keyString)

	if err != nil {
		return reqData, errors.Join(ErrInvalidKey, err)
	}

	expiration, err := g.getSecretExpiration(r)
	if err != nil {
		return reqData, err
	}

	maxReads, err := g.getSecretMaxReads(r)
	if err != nil {
		return reqData, err
	}

	reqData.UUID = UUID.String()
	reqData.Key = keyByte
	reqData.Expiration = expiration
	reqData.MaxReads = maxReads

	return reqData, nil
}
