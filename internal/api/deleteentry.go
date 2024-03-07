package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Ajnasz/sekret.link/internal/parsers"
)

// DeleteEntryManager is the interface for deleting an entry
type DeleteEntryManager interface {
	DeleteEntry(ctx context.Context, UUID string, deleteKey string) error
}

// DeleteEntryView is the interface for the view that should be implemented to render the delete entry results
type DeleteEntryView interface {
	RenderDeleteEntry(w http.ResponseWriter, r *http.Request)
	RenderDeleteEntryError(w http.ResponseWriter, r *http.Request, err error)
}

// DeleteHandler is the handler for deleting an entry
type DeleteHandler struct {
	entryManager DeleteEntryManager
	view         DeleteEntryView
}

// NewDeleteHandler creates a new DeleteHandler instance
func NewDeleteHandler(entryManager DeleteEntryManager, view DeleteEntryView) DeleteHandler {
	return DeleteHandler{
		entryManager: entryManager,
		view:         view,
	}
}

func (d DeleteHandler) handle(w http.ResponseWriter, r *http.Request) error {
	// TODO move parsing into the parsers package
	UUID, _, deleteKey, err := parsers.ParseDeleteEntryPath(r.URL.Path)

	if err != nil {
		fmt.Println("parse error", err)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := d.entryManager.DeleteEntry(ctx, UUID, deleteKey); err != nil {
		return err
	}

	d.view.RenderDeleteEntry(w, r)
	return nil
}

// Handle handles the delete request
func (d DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if err := d.handle(w, r); err != nil {
		d.view.RenderDeleteEntryError(w, r, err)
	}
}
