package parsers

import (
	"errors"
	"net/http"

	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/google/uuid"
)

type GetEntryParser struct{}

func NewGetEntryParser() GetEntryParser {
	return GetEntryParser{}
}

type GetEntryRequestData struct {
	UUID      string
	KeyString string
	Key       key.Key
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

	keyByte, err := getEntryKeyByte(keyString)

	if err != nil {
		return reqData, errors.Join(ErrInvalidKey, err)
	}
	return GetEntryRequestData{UUID: UUID.String(), Key: *keyByte, KeyString: keyString}, nil
}
