package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	apientries "github.com/Ajnasz/sekret.link/api/entries"
	"github.com/Ajnasz/sekret.link/encrypter/aes"
	"github.com/Ajnasz/sekret.link/entries"
	"github.com/Ajnasz/sekret.link/storage"
	"github.com/Ajnasz/sekret.link/storage/secret"
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

func handleGetSecret(entryStorage storage.Verifyable, UUID, keyString string) (*entries.Entry, error) {
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	secretStore := secret.NewSecretStorage(entryStorage, aes.New(key))
	ctx := context.Background()
	return secretStore.GetAndDelete(ctx, UUID)
}

func sendGetSecretResponse(entry *entries.Entry, keyString string, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") == "application/json" {
		response := apientries.SecretResponseFromEntryMeta(entry.EntryMeta)

		response.Data = string(entry.Data)
		response.Key = keyString
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(entry.Data)
	}
}

func handleGetEntry(entryStorage storage.Verifyable, w http.ResponseWriter, r *http.Request) {
	UUID, keyString, err := uuid.GetUUIDAndSecretFromPath(r.URL.Path)

	if err != nil {
		log.Println(err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err != nil {
		onGetError(w, err)
		return
	}

	entry, err := handleGetSecret(entryStorage, UUID, keyString)
	if err != nil {
		onGetError(w, err)
		return
	}

	sendGetSecretResponse(entry, keyString, w, r)
}
