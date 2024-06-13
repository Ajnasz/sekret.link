package views

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/uuid"
)

type EntryCreatedResponse struct {
	UUID      string
	Key       string
	Created   time.Time
	Accessed  time.Time
	Expire    time.Time
	DeleteKey string
}

func BuildCreatedResponse(meta *services.EntryMeta, keyString string) EntryCreatedResponse {
	return EntryCreatedResponse{
		UUID:      meta.UUID,
		Created:   meta.Created,
		Expire:    meta.Expire,
		Accessed:  meta.Accessed,
		DeleteKey: meta.DeleteKey,
		Key:       keyString,
	}
}

type EntryCreateView struct {
	webExternalURL *url.URL
}

func NewEntryCreateView(webExternalURL *url.URL) EntryCreateView {
	return EntryCreateView{webExternalURL: webExternalURL}
}

func (e EntryCreateView) Render(w http.ResponseWriter, r *http.Request, entry EntryCreatedResponse) {
	w.Header().Add("x-entry-uuid", entry.UUID)
	w.Header().Add("x-entry-key", entry.Key)
	w.Header().Add("x-entry-expire", entry.Expire.Format(time.RFC3339))
	w.Header().Add("x-entry-delete-key", entry.DeleteKey)

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(entry); err != nil {
			slog.Error("JSON encode failed", "error", err)
		}
	} else {
		newURL, err := uuid.GetUUIDUrlWithSecret(e.webExternalURL, entry.UUID, entry.Key)
		if err != nil {
			slog.Error("Get UUID URL with secret failed", "error", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", newURL.String())
	}
}

func (e EntryCreateView) RenderError(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Fprintf(os.Stderr, "Render error: %s", err)
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
