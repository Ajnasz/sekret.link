package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Ajnasz/sekret.link/aesencrypter"
	"github.com/Ajnasz/sekret.link/entries"
	"github.com/Ajnasz/sekret.link/key"
	"github.com/Ajnasz/sekret.link/storage"
	"github.com/Ajnasz/sekret.link/uuid"
)

func handleParseError(w http.ResponseWriter, err error) {
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
	} else if err.Error() == "Invalid data" {
		log.Println(err)
		http.Error(w, "Invalid max read", http.StatusBadRequest)
	} else {
		log.Println("Request parse error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

type CreateHandler struct {
	config HandlerConfig
}

func NewCreateHandler(c HandlerConfig) *CreateHandler {
	return &CreateHandler{c}
}

func (c CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, c.config.MaxDataSize)

	data, err := c.parseCreateRequest(r)

	if err != nil {
		handleParseError(w, err)
		return
	}

	k, err := key.NewGeneratedKey()

	if err != nil {
		log.Println("Create key failed", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	secretStore := storage.NewSecretStorage(c.config.EntryStorage, aesencrypter.New(k.Get()))

	UUID := uuid.NewUUIDString()

	err = secretStore.Create(UUID, data.body, data.expiration, data.maxReads)

	if err != nil {
		log.Println("Create secret failed", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	entry, err := secretStore.GetMeta(UUID)
	if err != nil {
		log.Println("Getting meta failed", err, UUID)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("x-entry-uuid", UUID)
	w.Header().Add("x-entry-key", k.ToHex())
	w.Header().Add("x-entry-expire", entry.Expire.Format(time.RFC3339))
	w.Header().Add("x-entry-delete-key", entry.DeleteKey)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")

		response := entries.SecretResponseFromEntryMeta(*entry)
		response.Key = k.ToHex()

		json.NewEncoder(w).Encode(response)
	} else {
		newURL, err := uuid.GetUUIDUrlWithSecret(c.config.WebExternalURL, UUID, k.ToHex())
		if err != nil {
			log.Println("Get UUID URL with secret failed", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", newURL.String())
	}
}
