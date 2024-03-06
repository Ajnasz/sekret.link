package parsers

import (
	"errors"
	"fmt"
	"path"

	"github.com/google/uuid"
)

// ErrInvalidURL is returned when the URL is invalid
var ErrInvalidURL = fmt.Errorf("invalid URL")

// ErrInvalidUUID is returned when the UUID is invalid
var ErrInvalidUUID = errors.New("invalid UUID")

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

	fmt.Println(pathDir)
	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return "", "", "", errors.Join(ErrInvalidUUID, err)
	}

	return UUID.String(), keyPart, delKey, nil
}
