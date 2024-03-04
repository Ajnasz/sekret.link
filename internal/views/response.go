package views

import (
	"errors"
	"log"
	"net/http"

	"github.com/Ajnasz/sekret.link/internal/parsers"
)

var ErrCreateKey = errors.New("create key failed")

func RenderErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	if err.Error() == "http: request body too large" {
		log.Println(err)
		http.Error(w, "Too large", http.StatusRequestEntityTooLarge)
	} else if errors.Is(err, parsers.ErrInvalidExpirationDate) {
		log.Println(err)
		http.Error(w, "Invalid expiration", http.StatusBadRequest)
		return
	} else if errors.Is(err, parsers.ErrInvalidMaxRead) {
		log.Println(err)
		http.Error(w, "Invalid max read", http.StatusBadRequest)
		return
	} else if errors.Is(err, parsers.ErrInvalidData) {
		log.Println(err)
		http.Error(w, "Invalid data", http.StatusBadRequest)
	} else {
		log.Println("Request parse error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}
