package api

import (
	"log"
	"net/http"
	"net/url"

	"github.com/Ajnasz/sekret.link/storage"
)

// HandlerConfig configuration for http handlers
type HandlerConfig struct {
	ExpireSeconds    int
	MaxExpireSeconds int
	EntryStorage     storage.VerifyConfirmReader
	MaxDataSize      int64
	WebExternalURL   *url.URL
}

// NewSecretHandler creates a SecretHandler instance
func NewSecretHandler(config HandlerConfig) SecretHandler {
	return SecretHandler{config}
}

// SecretHandler is an http.Handler implementation which handles requests to
// encode or decode the post body
type SecretHandler struct {
	config HandlerConfig
}

func (s SecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	if r.Method == http.MethodPost {
		if r.URL.Path != "/" && r.URL.Path != "" {
			http.Error(w, "Not found", http.StatusNotFound)
			log.Println("Not found", r.URL.Path)
			return
		}
		NewCreateHandler(s.config).Handle(w, r)
	} else if r.Method == http.MethodGet {
		handleGetEntry(s.config.EntryStorage, w, r)
	} else if r.Method == http.MethodDelete {
		handleDeleteEntry(s.config.EntryStorage, w, r)
	} else {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}
