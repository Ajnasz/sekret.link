package parsers

import (
	"errors"
	"path"

	"github.com/google/uuid"
)

func ParseDeleteEntryPath(urlPath string) (string, string, string, error) {
	// TODO pathdir might not exists if no webExternalURL is provided
	// fix that case
	pathDir, delKey := path.Split(urlPath)
	if len(pathDir) < 1 {
		return "", "", "", ErrInvalidURL
	}

	pathDir = pathDir[0 : len(pathDir)-1]

	uuidPart, keyPart := path.Split(pathDir)
	_, uuidFromPath := path.Split(uuidPart[0 : len(uuidPart)-1])

	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return "", "", "", errors.Join(ErrInvalidUUID, err)
	}

	return UUID.String(), keyPart, delKey, nil
}
