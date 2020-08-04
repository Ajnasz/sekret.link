package main

import (
	"encoding/hex"
	"log"
	"net/http"
)

func handleGetEntry(w http.ResponseWriter, r *http.Request) {
	UUID, keyString, err := getUUIDAndSecretFromPath(r.URL.Path)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	key, err := hex.DecodeString(keyString)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	secretStorage := &SecretStorage{storage, &AESEncrypter{key}}
	entry, err := secretStorage.GetAndDelete(UUID)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if len(entry) == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(entry)
}
