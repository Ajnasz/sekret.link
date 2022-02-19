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
	EntryStorage     storage.VerifyStorage
	MaxDataSize      int64
	WebExternalURL   *url.URL
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	if req.Header.Get("ORIGIN") != "" {
		(*w).Header().Set("Access-Control-Allow-Origin", req.Header.Get("ORIGIN"))
		(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	}
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
		setupResponse(&w, r)
		return
	}

	if r.Method == http.MethodPost {
		setupResponse(&w, r)
		if r.URL.Path != "/" && r.URL.Path != "" {
			http.Error(w, "Not found", http.StatusNotFound)
			log.Println("Not found", r.URL.Path)
			return
		}
		NewCreateHandler(s.config).Handle(w, r)
	} else if r.Method == http.MethodGet {
		setupResponse(&w, r)
		handleGetEntry(s.config.EntryStorage, w, r)
	} else if r.Method == http.MethodDelete {
		setupResponse(&w, r)
		handleDeleteEntry(s.config.EntryStorage, w, r)
	} else {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}
