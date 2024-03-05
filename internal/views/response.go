package views

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Ajnasz/sekret.link/internal/parsers"
)

var ErrCreateKey = errors.New("create key failed")

func RenderErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(err)
	if errors.Is(err, parsers.ErrInvalidExpirationDate) {
		http.Error(w, "Invalid expiration", http.StatusBadRequest)
		return
	} else if errors.Is(err, parsers.ErrInvalidMaxRead) {
		http.Error(w, "Invalid max read", http.StatusBadRequest)
		return
	} else if errors.Is(err, parsers.ErrInvalidData) {
		http.Error(w, "Invalid data", http.StatusBadRequest)
	} else if strings.Contains(err.Error(), "http: request body too large") {
		http.Error(w, "Too large", http.StatusRequestEntityTooLarge)
	} else {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}
