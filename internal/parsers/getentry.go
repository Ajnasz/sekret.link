package parsers

import (
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type GetEntryParser struct{}

func NewGetEntryParser() GetEntryParser {
	return GetEntryParser{}
}

type GetEntryRequestData struct {
	UUID      string
	KeyString string
	Key       []byte
}

func (g GetEntryParser) Parse(req *http.Request) (GetEntryRequestData, error) {
	var reqData GetEntryRequestData
	keyString := req.PathValue("key")
	if keyString == "" {
		return reqData, ErrInvalidKey
	}

	uuidFromPath := req.PathValue("uuid")
	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return reqData, errors.Join(ErrInvalidUUID, err)
	}
	key, err := hex.DecodeString(keyString)

	if err != nil {
		return reqData, errors.Join(ErrInvalidKey, err)
	}
	return GetEntryRequestData{UUID: UUID.String(), Key: key, KeyString: keyString}, nil
}
