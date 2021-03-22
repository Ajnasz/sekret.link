package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Ajnasz/sekret.link/storage"
)

func onGetError(w http.ResponseWriter, err error) {
	if err == ErrEntryExpired {
		http.Error(w, "Gone", http.StatusGone)
		return
	}

	if err == ErrEntryNotFound {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	log.Println(err)
	http.Error(w, "Internal error", http.StatusInternalServerError)
}

func handleGetSecret(UUID, keyString string) (*storage.Entry, error) {
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	secretStore := &secretStorage{entryStorage, &AESEncrypter{key}}
	return secretStore.GetAndDelete(UUID)
}

func sendGetSecretResponse(entry *storage.Entry, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") == "application/json" {
		response := secretResponseFromEntryMeta(entry.EntryMeta)
		response.Data = string(entry.Data)
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(entry.Data)
	}
}

func handleGetEntry(w http.ResponseWriter, r *http.Request) {
	UUID, keyString, err := getUUIDAndSecretFromPath(r.URL.Path)

	if err != nil {
		log.Println(err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	entry, err := handleGetSecret(UUID, keyString)

	if err != nil {
		onGetError(w, err)
		return
	}

	sendGetSecretResponse(entry, w, r)
}
