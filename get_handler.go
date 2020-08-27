package main

import (
	"encoding/hex"
	"encoding/json"
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
		if _, ok := err.(*entryExpiredError); ok {
			log.Println(err)
			http.Error(w, "Gone", http.StatusGone)
			return
		}

		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if len(entry.Data) == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		response := secretResponseFromEntryMeta(&entry.EntryMeta)

		response.Data = string(entry.Data)
		response.Key = keyString
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(entry.Data)
	}
}
