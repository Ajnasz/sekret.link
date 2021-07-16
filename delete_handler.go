package main

import (
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/Ajnasz/sekret.link/aesencrypter"
	"github.com/Ajnasz/sekret.link/storage"
	"github.com/google/uuid"
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

func handleDeleteEntry(entryStorage storage.VerifyStorage, w http.ResponseWriter, r *http.Request) {
	UUID, _, deleteKey, err := parseDeleteEntryPath(r.URL.Path)

	if err != nil {
		log.Println("Request parse error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	secretStore := storage.NewSecretStorage(entryStorage, aesencrypter.New(nil))

	validDeleteKey, err := secretStore.VerifyDelete(UUID, deleteKey)

	if err != nil {
		log.Println("Verifying delete failed", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !validDeleteKey {
		log.Println("Invalid delete key")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = secretStore.Delete(UUID)

	if err != nil {
		log.Println("Delete failed", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
