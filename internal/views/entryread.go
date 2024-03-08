package views

import (
	"crypto/aes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
)

type SecretResponse struct {
	UUID      string
	Key       string
	Data      string
	Created   time.Time
	Accessed  time.Time
	Expire    time.Time
	DeleteKey string
}

func buildSecretResponse(meta services.Entry) SecretResponse {
	return SecretResponse{
		UUID:      meta.UUID,
		Created:   meta.Created,
		Expire:    meta.Expire,
		Accessed:  meta.Accessed,
		DeleteKey: meta.DeleteKey,
		Data:      string(meta.Data),
	}
}

type EntryReadView struct{}

func NewEntryReadView() EntryReadView {
	return EntryReadView{}
}

func (e EntryReadView) RenderReadEntry(w http.ResponseWriter, r *http.Request, entry *services.Entry, keyString string) {
	if r.Header.Get("Accept") == "application/json" {
		response := buildSecretResponse(*entry)

		response.Key = keyString
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(entry.Data)
	}
}

func (e EntryReadView) RenderReadEntryError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, services.ErrEntryExpired) {
		http.Error(w, "Gone", http.StatusGone)
		return
	}

	if errors.Is(err, services.ErrEntryNotFound) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if errors.Is(err, parsers.ErrInvalidUUID) {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if errors.Is(err, parsers.ErrInvalidKey) {
		http.Error(w, "Bad request", http.StatusBadRequest)
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
