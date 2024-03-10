package uuid

import (
	"fmt"
	"net/url"
	"path"

	"github.com/google/uuid"
)

// GetUUIDUrlWithSecret is a formatter function which creates a path for a secret id and it's key
func GetUUIDUrlWithSecret(u *url.URL, UUID string, key string) (*url.URL, error) {
	newURL, err := url.Parse(fmt.Sprintf("%s/%s/%s", u.String(), UUID, key))
	if err != nil {
		return nil, err
	}

	return newURL, nil
}

// GetUUIDAndSecretFromPath extracts the secret id and it's key from a path
// url is expected to be in the format of (/.*)?/UUID/key
func GetUUIDAndSecretFromPath(urlPath string) (string, string, error) {
	// delete key is the last item from the path
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

// NewUUIDString Generates an uuid and returns as a string
func NewUUIDString() string {

	newUUID := uuid.New()

	return newUUID.String()
}
