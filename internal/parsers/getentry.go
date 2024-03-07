package parsers

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"path"

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

func (g GetEntryParser) Parse(u *http.Request) (GetEntryRequestData, error) {
	urlPath := u.URL.Path
	pathDir, keyString := path.Split(urlPath)
	var reqData GetEntryRequestData
	if len(pathDir) < 1 {
		return reqData, ErrInvalidURL
	}
	_, uuidFromPath := path.Split(pathDir[0 : len(pathDir)-1])
	UUID, err := uuid.Parse(uuidFromPath)

	fmt.Println("UUID ERROR", UUID, err)

	if err != nil {
		return reqData, errors.Join(ErrInvalidUUID, err)
	}
	key, err := hex.DecodeString(keyString)

	if err != nil {
		return reqData, err
	}
	return GetEntryRequestData{UUID: UUID.String(), Key: key, KeyString: keyString}, nil
}
