package api

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
	"github.com/Ajnasz/sekret.link/uuid"
)

var ErrInvalidUUIDError = errors.New("invalid UUID")
var ErrInvalidKeyError = errors.New("invalid key")

type GetHandler struct {
	DB *sql.DB
}

func (g GetHandler) handle(w http.ResponseWriter, r *http.Request) error {
	UUID, keyString, err := uuid.GetUUIDAndSecretFromPath(r.URL.Path)

	if err != nil {
		return ErrInvalidUUIDError
	}
	key, err := hex.DecodeString(keyString)

	if err != nil {
		return errors.Join(ErrInvalidKeyError, err)
	}

	encrypter := services.NewEncrypter(key)
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

func (g GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	err := g.handle(w, r)
	if err != nil {
		views.RenderReadEntryError(w, r, err)
	}
}
