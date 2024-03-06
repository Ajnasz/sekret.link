// Package api contains the http.Handler implementations for the api endpoints
package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/key"
)

// ErrInvalidExpirationDate request parse error happens when the user set
// expiration date is larger than the system maximum expiration date
var ErrInvalidExpirationDate = errors.New("Invalid expiration date")

// ErrInvalidMaxRead request parse error happens when the user maximum read
// number is greater than the system maximum read number
var ErrInvalidMaxRead = errors.New("Invalid max read")

// ErrInvalidData request parse error happens if the post data can not be accepted
var ErrInvalidData = errors.New("Invalid data")

// ErrRequestParseError request parse error happens if the post data can not be accepted
var ErrRequestParseError = errors.New("request parse error")

// CreateEntryParser is an interface for parsing the create entry request
type CreateEntryParser interface {
	Parse(r *http.Request) (*parsers.CreateEntryRequestData, error)
}

// CreateEntryManager is an interface for creating entries
type CreateEntryManager interface {
	CreateEntry(ctx context.Context, body []byte, maxReads int, expiration time.Duration) (*services.EntryMeta, error)
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
	key          *key.Key
}

// NewCreateHandler creates a new CreateHandler
func NewCreateHandler(
	maxDataSize int64,
	parser CreateEntryParser,
	entryManager CreateEntryManager,
	view CreateEntryView,
	key *key.Key,
) CreateHandler {
	return CreateHandler{
		maxDataSize:  maxDataSize,
		parser:       parser,
		entryManager: entryManager,
		view:         view,
		key:          key,
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
	entry, err := c.entryManager.CreateEntry(ctx, data.Body, data.MaxReads, data.Expiration)

	if err != nil {
		return err
	}

	c.view.RenderEntryCreated(w, r, entry, c.key.String())
	return nil
}

// Handle handles http request to create secret
func (c CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	err := c.handle(w, r)

	if err != nil {
		c.view.RenderCreateEntryErrorResponse(w, r, err)
	}
}
