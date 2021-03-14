package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func createKey() ([]byte, string, error) {
	key, err := generateRSAKey()
	if err != nil {
		return nil, "", err
	}

	keyString := hex.EncodeToString(key)

	return key, keyString, nil
}

func handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxDataSize)

	data, err := parseCreateRequest(r)

	if err != nil {
		if err.Error() == "http: request body too large" {
			http.Error(w, "Too large", http.StatusRequestEntityTooLarge)
		} else if err.Error() == "Invalid expiration date" {
			log.Println(err)
			http.Error(w, "Invalid expiration", http.StatusBadRequest)
			return
		} else if err.Error() == "Invalid max read" {
			log.Println(err)
			http.Error(w, "Invalid max read", http.StatusBadRequest)
			return
		} else {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}

	key, keyString, err := createKey()

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	secretStore := &secretStorage{entryStorage, &AESEncrypter{key}}

	UUID := newUUIDString()

	err = secretStore.Create(UUID, data.body, data.expiration, data.maxReads)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("x-entry-uuid", UUID)
	w.Header().Add("x-entry-key", keyString)
	w.Header().Add("x-entry-expire", time.Now().Add(data.expiration).Format(time.RFC3339))
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		entry, err := secretStore.GetMeta(UUID)

		if err != nil {
			log.Println(err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		response := secretResponseFromEntryMeta(*entry)
		response.Key = keyString

		json.NewEncoder(w).Encode(response)
	} else {
		newURL, err := getUUIDUrlWithSecret(webExternalURL, UUID, keyString)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", newURL.String())
	}
}
