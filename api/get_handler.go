package api

import (
	"context"
	"crypto/aes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	apientries "github.com/Ajnasz/sekret.link/api/entries"
	aesencrypter "github.com/Ajnasz/sekret.link/encrypter/aes"
	"github.com/Ajnasz/sekret.link/entries"
	"github.com/Ajnasz/sekret.link/storage"
	"github.com/Ajnasz/sekret.link/storage/secret"
	"github.com/Ajnasz/sekret.link/uuid"
)

func onGetError(w http.ResponseWriter, err error) {
	log.Println(err)
	if errors.Is(err, entries.ErrEntryExpired) {
		http.Error(w, "Gone", http.StatusGone)
		return
	}

	if errors.Is(err, entries.ErrEntryNotFound) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if errors.Is(err, hex.ErrLength) {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var keysizeError *aes.KeySizeError
	if errors.As(err, &keysizeError) {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	http.Error(w, "Internal error", http.StatusInternalServerError)
}

func handleGetSecret(entryStorage storage.VerifyConfirmReader, UUID, keyString string) (*entries.Entry, error) {
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, fmt.Errorf("hex decode error: %w", err)
	}

	secretStore := secret.NewSecretStorage(entryStorage, aesencrypter.New(key))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return secretStore.Read(ctx, UUID)
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

func handleGetEntry(entryStorage storage.VerifyConfirmReader, w http.ResponseWriter, r *http.Request) {
	UUID, keyString, err := uuid.GetUUIDAndSecretFromPath(r.URL.Path)

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
