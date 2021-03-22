package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Ajnasz/sekret.link/storage"
)

func createKey() ([]byte, string, error) {
	key, err := generateRSAKey()
	if err != nil {
		return nil, "", err
	}

	keyString := hex.EncodeToString(key)

	return key, keyString, nil
}

func handleReqestDataError(err error, w http.ResponseWriter) {
	log.Println(err)

	if err.Error() == "http: request body too large" {
		http.Error(w, "Too large", http.StatusRequestEntityTooLarge)
	} else if err.Error() == "Invalid expiration date" {
		http.Error(w, "Invalid expiration", http.StatusBadRequest)
		return
	} else if err.Error() == "Invalid max read" {
		http.Error(w, "Invalid max read", http.StatusBadRequest)
		return
	} else {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func createEntry(data *requestData) (*storage.EntryMeta, error) {
	key, keyString, err := createKey()

	if err != nil {
		return nil, err
	}

	secretStore := &secretStorage{entryStorage, &AESEncrypter{key}}

	UUID := newUUIDString()

	err = secretStore.Create(UUID, data.body, data.expiration, data.maxReads)

	if err != nil {
		return nil, err
	}

	entry, err := secretStore.GetMeta(UUID)
	if err != nil {
		return nil, err
	}

	entry.Key = keyString

	return entry, nil
}

func addEntryHeaders(entry *storage.EntryMeta, w http.ResponseWriter) {
	w.Header().Add("x-entry-uuid", entry.UUID)
	w.Header().Add("x-entry-key", entry.Key)
	w.Header().Add("x-entry-expire", entry.Expire.Format(time.RFC3339))
	w.Header().Add("x-entry-delete-key", entry.DeleteKey)
}

func sendCreateSecretResponse(entry *storage.EntryMeta, w http.ResponseWriter, r *http.Request) {
	addEntryHeaders(entry, w)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		response := secretResponseFromEntryMeta(*entry)
		response.Key = entry.Key
		json.NewEncoder(w).Encode(response)
	} else {
		newURL, err := getUUIDUrlWithSecret(webExternalURL, entry.UUID, entry.Key)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", newURL.String())
	}
}

func handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxDataSize)

	data, err := parseCreateRequest(r)

	if err != nil {
		handleReqestDataError(err, w)
		return
	}

	entry, err := createEntry(data)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	sendCreateSecretResponse(entry, w, r)

}
