package views

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/uuid"
)

type GenerateEntryKeyResponseData struct {
	// The UUID of the entry.
	UUID string
	// The key decryption key string of the entry.
	Key string

	// The time when the entry was created.
	Expire time.Time
}

// GenerateEntryKeyView is the view for the GenerateEntryKey endpoint.
type GenerateEntryKeyView struct {
	webExternalURL *url.URL
}

func NewGenerateEntryKeyView(webExternalURL *url.URL) GenerateEntryKeyView {
	return GenerateEntryKeyView{webExternalURL: webExternalURL}
}

// RenderGenerateEntryKey renders the response for the GenerateEntryKey endpoint.
func (g GenerateEntryKeyView) RenderGenerateEntryKey(w http.ResponseWriter, r *http.Request, response GenerateEntryKeyResponseData) {
	w.Header().Add("x-entry-uuid", response.UUID)
	w.Header().Add("x-entry-key", response.Key)
	w.Header().Add("x-entry-expire", response.Expire.Format(time.RFC3339))

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(response)
	} else {
		newURL, err := uuid.GetUUIDUrlWithSecret(g.webExternalURL, response.UUID, response.Key)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, newURL.String())
	}
}

// RenderGenerateEntryKeyError renders the error response for the GenerateEntryKey endpoint.
func (v GenerateEntryKeyView) RenderGenerateEntryKeyError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, parsers.ErrInvalidUUID) {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	} else if errors.Is(err, parsers.ErrInvalidKey) {
		http.Error(w, "Invalid key", http.StatusBadRequest)
		return
	} else if errors.Is(err, parsers.ErrInvalidExpirationDate) {
		http.Error(w, "Invalid expiration", http.StatusBadRequest)
		return
	} else if errors.Is(err, parsers.ErrInvalidMaxRead) {
		http.Error(w, "Invalid max read", http.StatusBadRequest)
		return
	} else {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}
