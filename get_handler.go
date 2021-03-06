package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Ajnasz/sekret.link/aesencrypter"
	"github.com/Ajnasz/sekret.link/entries"
	"github.com/Ajnasz/sekret.link/storage"
	"github.com/Ajnasz/sekret.link/uuid"
)

func onGetError(w http.ResponseWriter, err error) {
	if err == entries.ErrEntryExpired {
		http.Error(w, "Gone", http.StatusGone)
		return
	}

	if err == entries.ErrEntryNotFound {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	log.Println(err)
	http.Error(w, "Internal error", http.StatusInternalServerError)
}

func handleGetEntry(entryStorage storage.VerifyStorage, w http.ResponseWriter, r *http.Request) {
	UUID, keyString, err := uuid.GetUUIDAndSecretFromPath(r.URL.Path)

	if err != nil {
		log.Println(err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	key, err := hex.DecodeString(keyString)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	secretStore := storage.NewSecretStorage(entryStorage, aesencrypter.New(key))
	entry, err := secretStore.GetAndDelete(UUID)

	if err != nil {
		onGetError(w, err)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		response := entries.SecretResponseFromEntryMeta(entry.EntryMeta)

		response.Data = string(entry.Data)
		response.Key = keyString
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(entry.Data)
	}
}
