package api

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"

	"github.com/Ajnasz/sekret.link/internal/api"
	"github.com/Ajnasz/sekret.link/storage"
)

// HandlerConfig configuration for http handlers
type HandlerConfig struct {
	ExpireSeconds    int
	MaxExpireSeconds int
	EntryStorage     storage.VerifyConfirmReader
	MaxDataSize      int64
	WebExternalURL   *url.URL
	DB               *sql.DB
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

		createHandler := api.CreateHandler{
			MaxDataSize:      s.config.MaxDataSize,
			MaxExpireSeconds: s.config.MaxExpireSeconds,
			WebExternalURL:   s.config.WebExternalURL,
			DB:               s.config.DB,
		}
		createHandler.Handle(w, r)
		// NewCreateHandler(s.config).Handle(w, r)
	} else if r.Method == http.MethodGet {
		getHandler := api.GetHandler{
			DB: s.config.DB,
		}
		getHandler.Handle(w, r)
		// NewGetHandler(s.config).Handle(w, r)
	} else if r.Method == http.MethodDelete {
		deleteHandler := api.DeleteHandler{
			DB: s.config.DB,
		}
		deleteHandler.Handle(w, r)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
