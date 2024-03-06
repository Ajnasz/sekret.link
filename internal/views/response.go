package views

import (
	"crypto/aes"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Ajnasz/sekret.link/entries"
	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/google/uuid"
)

var ErrCreateKey = errors.New("create key failed")

func (e EntryView) RenderCreateEntryErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("CREATE ENTRY ERROR", err)
	if errors.Is(err, parsers.ErrInvalidExpirationDate) {
		http.Error(w, "Invalid expiration", http.StatusBadRequest)
		return
	} else if errors.Is(err, parsers.ErrInvalidMaxRead) {
		http.Error(w, "Invalid max read", http.StatusBadRequest)
		return
	} else if errors.Is(err, parsers.ErrInvalidData) {
		http.Error(w, "Invalid data", http.StatusBadRequest)
	} else if strings.Contains(err.Error(), "http: request body too large") {
		http.Error(w, "Too large", http.StatusRequestEntityTooLarge)
	} else {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (e EntryView) RenderReadEntryError(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(err)
	if errors.Is(err, services.ErrEntryExpired) {
		http.Error(w, "Gone", http.StatusGone)
		return
	}

	if errors.Is(err, services.ErrEntryNotFound) {
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

func (e EntryView) RenderDeleteEntryError(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(err)

	if errors.Is(err, entries.ErrEntryNotFound) || errors.Is(err, models.ErrEntryNotFound) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if errors.Is(err, models.ErrInvalidKey) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	if uuid.IsInvalidLengthError(err) || errors.Is(err, parsers.ErrInvalidUUID) {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	http.Error(w, "Internal error", http.StatusInternalServerError)
}
