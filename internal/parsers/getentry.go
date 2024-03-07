package parsers

import (
	"fmt"
	"path"

	"github.com/google/uuid"
)

func ParseGetEntryPath(urlPath string) (string, string, error) {
	pathDir, key := path.Split(urlPath)
	if len(pathDir) < 1 {
		return "", "", fmt.Errorf("Invalid URL %q", urlPath)
	}
	_, uuidFromPath := path.Split(pathDir[0 : len(pathDir)-1])
	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return "", "", err
	}
	return UUID.String(), key, nil
}
