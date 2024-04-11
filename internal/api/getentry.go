package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
)

// GetEntryManager is the interface for getting an entry
type GetEntryManager interface {
	ReadEntry(ctx context.Context, UUID string, key []byte) (*services.Entry, error)
}

// GetEntryView is the interface for the view that should be implemented to render the get entry results
type GetEntryView interface {
	RenderReadEntry(w http.ResponseWriter, r *http.Request, entry *services.Entry, key string)
	RenderReadEntryError(w http.ResponseWriter, r *http.Request, err error)
}

// ErrInvalidKeyError is returned when the key is invalid
var ErrInvalidKeyError = errors.New("invalid key")

// GetHandler is the handler for getting an entry
type GetHandler struct {
	entryManager GetEntryManager
	view         views.View[views.EntryReadResponse]
	parser       parsers.Parser[parsers.GetEntryRequestData]
}

// NewGetHandler creates a new GetHandler instance
func NewGetHandler(
	parser parsers.Parser[parsers.GetEntryRequestData],
	entryManager GetEntryManager,
	view views.View[views.EntryReadResponse],
) GetHandler {
	return GetHandler{
		view:         view,
		parser:       parser,
		entryManager: entryManager,
	}
}

func (g GetHandler) handle(w http.ResponseWriter, r *http.Request) error {
	request, err := g.parser.Parse(r)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	entry, err := g.entryManager.ReadEntry(ctx, request.UUID, request.Key)
	if err != nil {
		return err
	}

	g.view.Render(w, r, views.BuildEntryReadResponse(*entry, request.KeyString))

	return nil
}

// Handle handles http request to get secret
func (g GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	err := g.handle(w, r)
	if err != nil {
		g.view.RenderError(w, r, err)
	}
}
