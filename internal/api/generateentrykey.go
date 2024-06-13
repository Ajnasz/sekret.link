package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
)

type GenerateEntryKeyView interface {
	RenderGenerateEntryKey(w http.ResponseWriter, r *http.Request, entry views.GenerateEntryKeyResponseData)
	RenderGenerateEntryKeyError(w http.ResponseWriter, r *http.Request, err error)
}

type GenerateEntryKeyManager interface {
	GenerateEntryKey(ctx context.Context, UUID string, k key.Key, expire *time.Duration, maxReads *int) (*services.EntryKeyData, error)
}

type GenerateEntryKeyHandler struct {
	entryManager GenerateEntryKeyManager
	view         views.View[views.GenerateEntryKeyResponseData]
	parser       parsers.Parser[parsers.GenerateEntryKeyRequestData]
}

func NewGenerateEntryKeyHandler(
	parser parsers.Parser[parsers.GenerateEntryKeyRequestData],
	entryManager GenerateEntryKeyManager,
	view views.View[views.GenerateEntryKeyResponseData],
) GenerateEntryKeyHandler {
	return GenerateEntryKeyHandler{
		view:         view,
		parser:       parser,
		entryManager: entryManager,
	}
}

func (g GenerateEntryKeyHandler) handle(w http.ResponseWriter, r *http.Request) error {
	request, err := g.parser.Parse(r)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	entry, err := g.entryManager.GenerateEntryKey(ctx, request.UUID, request.Key, &request.Expiration, &request.MaxReads)
	if err != nil {
		return err
	}

	g.view.Render(w, r, views.GenerateEntryKeyResponseData{
		UUID:   request.UUID,
		Key:    entry.KEK,
		Expire: entry.Expire,
	})
	return nil
}

func (g GenerateEntryKeyHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if err := g.handle(w, r); err != nil {
		g.view.RenderError(w, r, err)
	}
}
