package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"path"

	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
	"github.com/google/uuid"
)

var ErrInvalidURL = fmt.Errorf("invalid URL")

func parseDeleteEntryPath(urlPath string) (string, string, string, error) {
	pathDir, delKey := path.Split(urlPath)
	if len(pathDir) < 1 {
		return "", "", "", ErrInvalidURL
	}

	pathDir = pathDir[0 : len(pathDir)-1]

	uuidPart, keyPart := path.Split(pathDir)
	_, uuidFromPath := path.Split(uuidPart[0 : len(uuidPart)-1])

	fmt.Println(pathDir)
	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return "", "", "", err
	}

	return UUID.String(), keyPart, delKey, nil
}

type DeleteHandler struct {
	DB *sql.DB
}

func (d DeleteHandler) NewDeleteHandler() DeleteHandler {
	return DeleteHandler{}
}

func (d DeleteHandler) handle(w http.ResponseWriter, r *http.Request) error {
	UUID, _, deleteKey, err := parseDeleteEntryPath(r.URL.Path)

	if err != nil {
		return err
	}

	encrypter := services.NewEncrypter([]byte{})
	entryManager := services.NewEntryManager(d.DB, &models.EntryModel{}, encrypter)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := entryManager.DeleteEntry(ctx, UUID, deleteKey); err != nil {
		return err
	}

	views.RenderDeleteEntry(w, r)
	return nil
}
func (d DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	err := d.handle(w, r)
	if err != nil {
		views.RenderDeleteEntryError(w, r, err)
	}
}
