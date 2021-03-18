package main

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"path"
)

func parseDeleteEntryPath(urlPath string) (string, string, string, error) {
	pathDir, delKey := path.Split(urlPath)
	if len(pathDir) < 1 {
		return "", "", "", fmt.Errorf("Invalid URL %q", urlPath)
	}

	pathDir = pathDir[0 : len(pathDir)-1]

	uuidPart, keyPart := path.Split(pathDir)
	_, uuidFromPath := path.Split(uuidPart[0 : len(uuidPart)-1])
	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return "", "", "", err
	}

	return UUID.String(), keyPart, delKey, nil
}

func handleDeleteEntry(w http.ResponseWriter, r *http.Request) {
	UUID, _, deleteKey, err := parseDeleteEntryPath(r.URL.Path)

	if err != nil {
		log.Println(err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	secretStore := &secretStorage{entryStorage, &AESEncrypter{}}

	validDeleteKey, err := secretStore.VerifyDelete(UUID, deleteKey)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !validDeleteKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = secretStore.Delete(UUID)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
