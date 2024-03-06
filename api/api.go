package api

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"

	"github.com/Ajnasz/sekret.link/internal/api"
	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
	"github.com/Ajnasz/sekret.link/key"
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

		k, err := key.NewGeneratedKey()

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		parser := parsers.NewCreateEntryParser(s.config.MaxExpireSeconds)
		encrypter := services.NewAESEncrypter(k.Get())
		entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, encrypter)
		view := views.NewEntryView(s.config.WebExternalURL)

		createHandler := api.NewCreateHandler(
			s.config.MaxDataSize,
			parser,
			entryManager,
			view,
			k,
		)
		createHandler.Handle(w, r)
		// NewCreateHandler(s.config).Handle(w, r)
	} else if r.Method == http.MethodGet {
		view := views.NewEntryView(s.config.WebExternalURL)
		getHandler := api.NewGetHandler(
			s.config.DB,
			view,
		)
		getHandler.Handle(w, r)
		// NewGetHandler(s.config).Handle(w, r)
	} else if r.Method == http.MethodDelete {
		encrypter := services.NewAESEncrypter([]byte{})
		entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, encrypter)
		view := views.NewEntryView(s.config.WebExternalURL)
		deleteHandler := api.NewDeleteHandler(entryManager, view)
		deleteHandler.Handle(w, r)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
