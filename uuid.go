package main

import (
	"fmt"
	"net/url"
	"path"

	"github.com/google/uuid"
)

func getUUIDUrl(u *url.URL, UUID string) (*url.URL, error) {
	newURL, err := url.Parse(fmt.Sprintf("%s/%s", u.String(), UUID))
	if err != nil {
		return nil, err
	}

	return newURL, nil
}

func getUUIDUrlWithSecret(u *url.URL, UUID string, key string) (*url.URL, error) {
	newURL, err := url.Parse(fmt.Sprintf("%s/%s/%s", u.String(), UUID, key))
	if err != nil {
		return nil, err
	}

	return newURL, nil
}

func getUUIDAndSecretFromPath(urlPath string) (string, string, error) {
	pathDir, key := path.Split(urlPath)
	if len(pathDir) < 1 {
		return "", "", fmt.Errorf("Invalid URL")
	}
	_, uuidFromPath := path.Split(pathDir[0 : len(pathDir)-1])
	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return "", "", err
	}
	return UUID.String(), key, nil
}

func getUUIDFromPath(urlPath string) (string, error) {
	_, uuidFromPath := path.Split(urlPath)
	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return "", err
	}
	return UUID.String(), nil
}

func newUUIDString() string {

	newUUID := uuid.New()

	return newUUID.String()
}
