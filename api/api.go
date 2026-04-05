package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/Ajnasz/sekret.link/api/middlewares"
	"github.com/Ajnasz/sekret.link/internal/api"
	"github.com/Ajnasz/sekret.link/internal/hasher"
	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
)

func newAESEncrypter(b key.Key) services.Encrypter {
	return services.NewAESEncrypter(b)
}

// HandlerConfig configuration for http handlers
type HandlerConfig struct {
	ExpireSeconds    int
	MaxExpireSeconds int
	MaxDataSize      int64
	WebExternalURL   *url.URL
	DB               *sql.DB
	CORSOrigins      []string
}

// SecretHandler is an http.Handler implementation which handles requests to
// encode or decode the post body
type SecretHandler struct {
	config HandlerConfig
}

// NewSecretHandler creates a SecretHandler instance
func NewSecretHandler(config HandlerConfig) SecretHandler {
	return SecretHandler{config: config}
}

// POST method handler
// This method is responsible for creating a new entry
// url: /
// query:
//   - expire: the expiration time of the entry
//   - maxReads: the maximum number of reads for the entry
//
// method: POST
// response: 200 OK
// response: 400 Bad Request
// response: 500 Internal Server Error
// response: 413 Payload Too Large
func (s SecretHandler) Post(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "" {
			http.Error(w, "Not found", http.StatusNotFound)
			slog.Info("Not found", "url", r.URL.Path)
			return
		}

		parser := parsers.NewCreateEntryParser(s.config.MaxExpireSeconds)
		keyManager := services.NewEntryKeyManager(s.config.DB, &models.EntryKeyModel{}, hasher.NewSHA256Hasher(), newAESEncrypter)
		entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, newAESEncrypter, keyManager)
		view := views.NewEntryCreateView(s.config.WebExternalURL)

		createHandler := api.NewCreateHandler(
			s.config.MaxDataSize,
			parser,
			entryManager,
			view,
		)
		createHandler.Handle(ctx, w, r)
	}
}

// GET method handler
func (s SecretHandler) Get(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		view := views.NewEntryReadView()
		parser := parsers.NewGetEntryParser()
		keyManager := services.NewEntryKeyManager(s.config.DB, &models.EntryKeyModel{}, hasher.NewSHA256Hasher(), newAESEncrypter)
		entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, newAESEncrypter, keyManager)
		getHandler := api.NewGetHandler(
			parser,
			entryManager,
			view,
		)
		getHandler.Handle(ctx, w, r)
	}
}

// DELETE method handler
func (s SecretHandler) Delete(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		keyManager := services.NewEntryKeyManager(s.config.DB, &models.EntryKeyModel{}, hasher.NewSHA256Hasher(), newAESEncrypter)
		entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, newAESEncrypter, keyManager)
		view := views.NewEntryDeleteView()
		deleteHandler := api.NewDeleteHandler(entryManager, view)
		deleteHandler.Handle(w, r)
	}
}

// OPTIONS method handler
func (s SecretHandler) Options(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// GenerateEncryptionKey provides a way to generate a new encryption key for an existing entry
// This allows to share the same entry with multiple users without sharing the encryption key
// url: /key/{uuid}/{key}
// - uuid: the uuid of the entry
// - key: the key of the entry
// query:
//   - expire: the expiration time of the new key
//   - maxReads: the maximum number of reads for the new key
//
// method: GET
// response: 200 OK
func (s SecretHandler) GenerateEncryptionKey(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		keyManager := services.NewEntryKeyManager(s.config.DB, &models.EntryKeyModel{}, hasher.NewSHA256Hasher(), newAESEncrypter)
		entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, newAESEncrypter, keyManager)
		view := views.NewGenerateEntryKeyView(s.config.WebExternalURL)
		parser := parsers.NewGenerateEntryKeyParser(s.config.MaxExpireSeconds)
		getHandler := api.NewGenerateEntryKeyHandler(
			parser,
			entryManager,
			view,
		)

		getHandler.Handle(ctx, w, r)
	}
}

// NotFound handler
func (s SecretHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not found", http.StatusNotFound)
}

func (s SecretHandler) ServeHTTP(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.Post(ctx)(w, r)
		case http.MethodGet:
			s.Get(ctx)(w, r)
		case http.MethodDelete:
			s.Delete(ctx)(w, r)
		case http.MethodOptions:
			s.Options(ctx)(w, r)
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}
}

func clearApiRoot(apiRoot string) string {
	apiRoot = path.Clean(path.Join("/", apiRoot))

	if strings.HasSuffix(apiRoot, "/") {
		return apiRoot
	}

	return apiRoot + "/"
}

func (s SecretHandler) RegisterHandlers(ctx context.Context, mux *http.ServeMux, apiRoot string) {
	apiRoot = clearApiRoot(apiRoot)
	mux.Handle(
		fmt.Sprintf("GET %s", path.Join("/", apiRoot, "{uuid}", "{key}")),
		http.StripPrefix(
			apiRoot,
			middlewares.SetupLogging(
				false,
				middlewares.SetupHeaders(s.config.CORSOrigins, http.HandlerFunc(s.Get(ctx))),
			),
		),
	)
	mux.Handle(
		fmt.Sprintf("POST %s", apiRoot),
		http.StripPrefix(
			path.Join("/", apiRoot),
			middlewares.SetupLogging(
				true,
				middlewares.SetupHeaders(s.config.CORSOrigins, http.HandlerFunc(s.Post(ctx))),
			),
		),
	)

	mux.Handle(
		fmt.Sprintf("DELETE %s", path.Join("/", apiRoot, "{uuid}", "{key}", "{deleteKey}")),
		http.StripPrefix(
			apiRoot,
			middlewares.SetupLogging(
				false,
				middlewares.SetupHeaders(s.config.CORSOrigins, http.HandlerFunc(s.Delete(ctx))),
			),
		),
	)

	mux.Handle(
		fmt.Sprintf("OPTIONS %s", apiRoot),
		http.StripPrefix(
			apiRoot,
			middlewares.SetupLogging(
				false,
				middlewares.SetupHeaders(s.config.CORSOrigins, http.HandlerFunc(s.Options(ctx))),
			),
		),
	)

	// TODO
	mux.Handle(
		fmt.Sprintf("GET %s", path.Join(apiRoot, "key", "{uuid}", "{key}")),
		http.StripPrefix(
			apiRoot,
			middlewares.SetupLogging(
				false,
				middlewares.SetupHeaders(s.config.CORSOrigins, http.HandlerFunc(s.GenerateEncryptionKey(ctx))),
			),
		),
	)

	mux.Handle("/", middlewares.SetupLogging(true, middlewares.SetupHeaders(s.config.CORSOrigins, http.HandlerFunc(s.NotFound))))

}
