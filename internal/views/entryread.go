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

type EntryReadResponse struct {
	UUID      string
	Key       string
	Data      string
	Created   time.Time
	Accessed  time.Time
	Expire    time.Time
	DeleteKey string
}

func BuildEntryReadResponse(meta services.Entry, key string) EntryReadResponse {
	return EntryReadResponse{
		UUID:      meta.UUID,
		Key:       key,
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

func (e EntryReadView) Render(w http.ResponseWriter, r *http.Request, response EntryReadResponse) {
	if r.Header.Get("Accept") == "application/json" {
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response.Data))
	}
}

func (e EntryReadView) RenderError(w http.ResponseWriter, r *http.Request, err error) {
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
