// Package api contains the http.Handler implementations for the api endpoints
package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/url"

	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
	"github.com/Ajnasz/sekret.link/key"
)

// ErrInvalidExpirationDate request parse error happens when the user set
// expiration date is larger than the system maximum expiration date
var ErrInvalidExpirationDate = errors.New("Invalid expiration date")

// ErrInvalidMaxRead request parse error happens when the user maximum read
// number is greater than the system maximum read number
var ErrInvalidMaxRead = errors.New("Invalid max read")

// ErrInvalidData request parse error happens if the post data can not be accepted
var ErrInvalidData = errors.New("Invalid data")

// CreateHandler is an http.Handler implementaton which creates secrets
type CreateHandler struct {
	MaxDataSize      int64
	MaxExpireSeconds int
	WebExternalURL   *url.URL
	DB               *sql.DB
}

func (c CreateHandler) handle(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, c.MaxDataSize)

	parser := parsers.NewCreateEntryParser(c.MaxExpireSeconds)

	data, err := parser.Parse(r)

	if err != nil {
		return errors.Join(errors.New("request parse error"), err)
	}

	k, err := key.NewGeneratedKey()

	if err != nil {
		return errors.Join(errors.New("key generate failed"), err)
	}

	encrypter := services.NewEncrypter(k.Get())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entryManager := services.NewEntryManager(c.DB, &models.EntryModel{}, encrypter)
	entry, err := entryManager.CreateEntry(ctx, data.Body, data.MaxReads, data.Expiration)

	if err != nil {
		return errors.New("create secret failed")
	}

	view := views.NewEntryView(c.WebExternalURL)
	view.RenderEntryCreated(w, r, entry, k.String())
	return nil
}

// Handle handles http request to create secret
func (c CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	err := c.handle(w, r)

	if err != nil {
		views.RenderCreateEntryErrorResponse(w, r, err)
	}
}
