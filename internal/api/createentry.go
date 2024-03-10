package api

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
)

// CreateEntryParser is an interface for parsing the create entry request
type CreateEntryParser interface {
	Parse(r *http.Request) (*parsers.CreateEntryRequestData, error)
}

// CreateEntryManager is an interface for creating entries
type CreateEntryManager interface {
	CreateEntry(ctx context.Context, body []byte, maxReads int, expiration time.Duration) (*services.EntryMeta, []byte, error)
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
	view         CreateEntryView
}

// NewCreateHandler creates a new CreateHandler
func NewCreateHandler(
	maxDataSize int64,
	parser CreateEntryParser,
	entryManager CreateEntryManager,
	view CreateEntryView,
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
	entry, key, err := c.entryManager.CreateEntry(ctx, data.Body, data.MaxReads, data.Expiration)

	if err != nil {
		return err
	}

	c.view.RenderEntryCreated(w, r, entry, hex.EncodeToString(key))
	return nil
}

// Handle handles http request to create secret
func (c CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	err := c.handle(w, r)

	if err != nil {
		c.view.RenderCreateEntryErrorResponse(w, r, err)
	}
}
