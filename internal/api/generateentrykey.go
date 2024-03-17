package api

import (
	"context"
	"encoding/hex"
	"net/http"

	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
)

type GenerateEntryKeyView interface {
	RenderGenerateEntryKey(w http.ResponseWriter, r *http.Request, entry views.GenerateEntryKeyResponseData)
	RenderGenerateEntryKeyError(w http.ResponseWriter, r *http.Request, err error)
}

type GenerateEntryKeyManager interface {
	GenerateEntryKey(ctx context.Context, UUID string, key []byte) (*services.EntryKeyData, error)
}

type GenerateEntryKeyHandler struct {
	entryManager GenerateEntryKeyManager
	view         GenerateEntryKeyView
	parser       parsers.Parser[parsers.GenerateEntryKeyRequestData]
}

func NewGenerateEntryKeyHandler(
	parser parsers.Parser[parsers.GenerateEntryKeyRequestData],
	entryManager GenerateEntryKeyManager,
	view GenerateEntryKeyView,
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

	entry, err := g.entryManager.GenerateEntryKey(ctx, request.UUID, request.Key)
	if err != nil {
		return err
	}

	g.view.RenderGenerateEntryKey(w, r, views.GenerateEntryKeyResponseData{
		UUID:   request.UUID,
		Key:    hex.EncodeToString(entry.KEK),
		Expire: entry.Expire,
	})
	return nil
}

func (g GenerateEntryKeyHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if err := g.handle(w, r); err != nil {
		g.view.RenderGenerateEntryKeyError(w, r, err)
	}
}
