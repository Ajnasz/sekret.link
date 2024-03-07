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

	// encrypter := services.NewAESEncrypter(k.Get())
	encrypter := func(b []byte) services.Encrypter {
		return services.NewAESEncrypter(b)
	}
	if r.Method == http.MethodPost {
		if r.URL.Path != "/" && r.URL.Path != "" {
			http.Error(w, "Not found", http.StatusNotFound)
			log.Println("Not found", r.URL.Path)
			return
		}

		parser := parsers.NewCreateEntryParser(s.config.MaxExpireSeconds)
		entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, encrypter)
		view := views.NewEntryView(s.config.WebExternalURL)

		createHandler := api.NewCreateHandler(
			s.config.MaxDataSize,
			parser,
			entryManager,
			view,
		)
		createHandler.Handle(w, r)
		// NewCreateHandler(s.config).Handle(w, r)
	} else if r.Method == http.MethodGet {
		view := views.NewEntryView(s.config.WebExternalURL)
		parser := parsers.NewGetEntryParser()
		entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, encrypter)
		getHandler := api.NewGetHandler(
			parser,
			entryManager,
			view,
		)
		getHandler.Handle(w, r)
		// NewGetHandler(s.config).Handle(w, r)
	} else if r.Method == http.MethodDelete {
		entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, encrypter)
		view := views.NewEntryView(s.config.WebExternalURL)
		deleteHandler := api.NewDeleteHandler(entryManager, view)
		deleteHandler.Handle(w, r)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
