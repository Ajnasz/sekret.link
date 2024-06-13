package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
)

// CreateEntryParser is an interface for parsing the create entry request
type CreateEntryParser interface {
	Parse(r *http.Request) (*parsers.CreateEntryRequestData, error)
}

// CreateEntryManager is an interface for creating entries
type CreateEntryManager interface {
	CreateEntry(ctx context.Context, contentType string, body []byte, expiration *time.Duration, maxReads *int) (*services.EntryMeta, key.Key, error)
}

// CreateEntryView is an interface for rendering the create entry response
type CreateEntryView interface {
	RenderEntryCreated(w http.ResponseWriter, r *http.Request, entry *services.EntryMeta, key string)
	RenderCreateEntryErrorResponse(w http.ResponseWriter, r *http.Request, err error)
}

// CreateHandler is an http.Handler implementaton which creates secrets
type CreateHandler struct {
	maxDataSize  int64
	parser       CreateEntryParser
	entryManager CreateEntryManager
	view         views.View[views.EntryCreatedResponse]
}

// NewCreateHandler creates a new CreateHandler
func NewCreateHandler(
	maxDataSize int64,
	parser CreateEntryParser,
	entryManager CreateEntryManager,
	view views.View[views.EntryCreatedResponse],
) CreateHandler {
	return CreateHandler{
		maxDataSize:  maxDataSize,
		parser:       parser,
		entryManager: entryManager,
		view:         view,
	}
}

func (c CreateHandler) handle(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, c.maxDataSize)

	data, err := c.parser.Parse(r)

	if err != nil {
		return errors.Join(ErrRequestParseError, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entry, key, err := c.entryManager.CreateEntry(ctx, data.ContentType, data.Body, &data.Expiration, &data.MaxReads)

	if err != nil {
		return err
	}

	viewData := views.BuildCreatedResponse(entry, key.String())

	c.view.Render(w, r, viewData)
	return nil
}

// Handle handles http request to create secret
func (c CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if err := c.handle(w, r); err != nil {
		log.Println("create error", err)
		c.view.RenderError(w, r, err)
	}
}
