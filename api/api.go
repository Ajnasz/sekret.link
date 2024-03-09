package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/Ajnasz/sekret.link/api/middlewares"
	"github.com/Ajnasz/sekret.link/internal/api"
	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
)

// HandlerConfig configuration for http handlers
type HandlerConfig struct {
	ExpireSeconds    int
	MaxExpireSeconds int
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

// POST method handler
func (s SecretHandler) Post(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "" {
		http.Error(w, "Not found", http.StatusNotFound)
		log.Println("Not found", r.URL.Path)
		return
	}

	encrypter := func(b []byte) services.Encrypter {
		return services.NewAESEncrypter(b)
	}

	parser := parsers.NewCreateEntryParser(s.config.MaxExpireSeconds)
	entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, encrypter)
	view := views.NewEntryCreateView(s.config.WebExternalURL)

	createHandler := api.NewCreateHandler(
		s.config.MaxDataSize,
		parser,
		entryManager,
		view,
	)
	createHandler.Handle(w, r)
}

// GET method handler
func (s SecretHandler) Get(w http.ResponseWriter, r *http.Request) {
	encrypter := func(b []byte) services.Encrypter {
		return services.NewAESEncrypter(b)
	}

	view := views.NewEntryReadView()
	parser := parsers.NewGetEntryParser()
	entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, encrypter)
	getHandler := api.NewGetHandler(
		parser,
		entryManager,
		view,
	)
	getHandler.Handle(w, r)
}

// DELETE method handler
func (s SecretHandler) Delete(w http.ResponseWriter, r *http.Request) {
	encrypter := func(b []byte) services.Encrypter {
		return services.NewAESEncrypter(b)
	}

	entryManager := services.NewEntryManager(s.config.DB, &models.EntryModel{}, encrypter)
	view := views.NewEntryDeleteView()
	deleteHandler := api.NewDeleteHandler(entryManager, view)
	deleteHandler.Handle(w, r)
}

// OPTIONS method handler
func (s SecretHandler) Options(w http.ResponseWriter, r *http.Request) {
	// Your OPTIONS method logic goes here
	w.WriteHeader(http.StatusOK)
}

// NotFound handler
func (s SecretHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not found", http.StatusNotFound)
}

func (s SecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.Post(w, r)
	case http.MethodGet:
		s.Get(w, r)
	case http.MethodDelete:
		s.Delete(w, r)
	case http.MethodOptions:
		s.Options(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func clearApiRoot(apiRoot string) string {
	apiRoot = path.Clean(path.Join("/", apiRoot))

	if strings.HasSuffix(apiRoot, "/") {
		return apiRoot
	}

	return apiRoot + "/"
}

func (s SecretHandler) RegisterHandlers(mux *http.ServeMux, apiRoot string) {
	mux.Handle(
		fmt.Sprintf("GET %s", path.Join("/", apiRoot, "{uuid}", "{key}")),
		http.StripPrefix(
			apiRoot,
			middlewares.SetupLogging(
				middlewares.SetupHeaders(http.HandlerFunc(s.Get)),
			),
		),
	)
	mux.Handle(
		fmt.Sprintf("POST %s", clearApiRoot(apiRoot)),
		http.StripPrefix(
			path.Join("/", apiRoot),
			middlewares.SetupLogging(
				middlewares.SetupHeaders(http.HandlerFunc(s.Post)),
			),
		),
	)

	mux.Handle(
		fmt.Sprintf("DELETE %s", path.Join("/", apiRoot, "{uuid}", "{key}", "{deleteKey}")),
		http.StripPrefix(
			apiRoot,
			middlewares.SetupLogging(
				middlewares.SetupHeaders(http.HandlerFunc(s.Delete)),
			),
		),
	)

	mux.Handle(
		fmt.Sprintf("OPTIONS %s", clearApiRoot(apiRoot)),
		http.StripPrefix(
			apiRoot,
			middlewares.SetupLogging(
				middlewares.SetupHeaders(http.HandlerFunc(s.Options)),
			),
		),
	)

	mux.Handle("/", middlewares.SetupLogging(middlewares.SetupHeaders(http.HandlerFunc(s.NotFound))))

}
