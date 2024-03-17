package views

import (
	"errors"
	"net/http"

	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/google/uuid"
)

type DeleteEntryResponse struct{}

type EntryDeleteView struct{}

func NewEntryDeleteView() EntryDeleteView {
	return EntryDeleteView{}
}
func (e EntryDeleteView) Render(w http.ResponseWriter, r *http.Request, data DeleteEntryResponse) {
	w.WriteHeader(http.StatusAccepted)
}

func (e EntryDeleteView) RenderError(w http.ResponseWriter, r *http.Request, err error) {

	if errors.Is(err, models.ErrEntryNotFound) || errors.Is(err, models.ErrEntryNotFound) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if errors.Is(err, models.ErrInvalidKey) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	if uuid.IsInvalidLengthError(err) || errors.Is(err, parsers.ErrInvalidUUID) {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	http.Error(w, "Internal error", http.StatusInternalServerError)
}
