package api

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
	"github.com/Ajnasz/sekret.link/uuid"
)

// ErrInvalidKeyError is returned when the key is invalid
var ErrInvalidKeyError = errors.New("invalid key")

// GetHandler is the handler for getting an entry
type GetHandler struct {
	DB *sql.DB
}

func (g GetHandler) handle(w http.ResponseWriter, r *http.Request) error {
	// TODO move to parsers
	UUID, keyString, err := uuid.GetUUIDAndSecretFromPath(r.URL.Path)

	if err != nil {
		return parsers.ErrInvalidUUID
	}
	key, err := hex.DecodeString(keyString)

	if err != nil {
		return errors.Join(ErrInvalidKeyError, err)
	}

	encrypter := services.NewAESEncrypter(key)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entryManager := services.NewEntryManager(g.DB, &models.EntryModel{}, encrypter)

	entry, err := entryManager.ReadEntry(ctx, UUID)
	if err != nil {
		return err
	}

	views.RenderReadEntry(w, r, entry, keyString)

	return nil
}

// Handle handles http request to get secret
func (g GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	err := g.handle(w, r)
	if err != nil {
		views.RenderReadEntryError(w, r, err)
	}
}
