package views

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/uuid"
)

type EntryCreatedResponse struct {
	UUID      string
	Key       string
	Created   time.Time
	Accessed  time.Time
	Expire    time.Time
	DeleteKey string
}

func buildCreatedResponse(meta *services.EntryMeta, keyString string) EntryCreatedResponse {
	return EntryCreatedResponse{
		UUID:      meta.UUID,
		Created:   meta.Created,
		Expire:    meta.Expire,
		Accessed:  meta.Accessed,
		DeleteKey: meta.DeleteKey,
		Key:       keyString,
	}
}

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

type EntryView struct {
	WebExternalURL *url.URL
}

func NewEntryView(webExternalURL *url.URL) EntryView {
	return EntryView{
		WebExternalURL: webExternalURL,
	}
}

func (e EntryView) RenderEntryCreated(w http.ResponseWriter, r *http.Request, entry *services.EntryMeta, keyString string) {
	w.Header().Add("x-entry-uuid", entry.UUID)
	w.Header().Add("x-entry-key", keyString)
	w.Header().Add("x-entry-expire", entry.Expire.Format(time.RFC3339))
	w.Header().Add("x-entry-delete-key", entry.DeleteKey)

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")

		response := buildCreatedResponse(entry, keyString)

		json.NewEncoder(w).Encode(response)
	} else {
		newURL, err := uuid.GetUUIDUrlWithSecret(e.WebExternalURL, entry.UUID, keyString)
		if err != nil {
			log.Println("Get UUID URL with secret failed", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", newURL.String())
	}
}

func RenderReadEntry(w http.ResponseWriter, r *http.Request, entry *services.Entry, keyString string) {
	if r.Header.Get("Accept") == "application/json" {
		response := buildSecretResponse(*entry)

		response.Key = keyString
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(entry.Data)
	}
}

func (e EntryView) RenderDeleteEntry(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}
