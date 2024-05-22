package parsers

import (
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers/expiration"
	"github.com/Ajnasz/sekret.link/internal/parsers/maxreads"
)

type CreateEntryParser struct {
	maxExpireSeconds int
}

type CreateEntryRequestData struct {
	ContentType string
	Body        []byte
	Expiration  time.Duration
	MaxReads    int
}

func NewCreateEntryParser(maxExpireSeconds int) CreateEntryParser {
	return CreateEntryParser{maxExpireSeconds: maxExpireSeconds}
}

func parseMultiForm(r *http.Request) ([]byte, string, error) {
	err := r.ParseMultipartForm(1024 * 1024)
	if err != nil {
		return nil, "", err
	}

	secret := r.PostForm.Get("secret")
	if secret != "" {
		body := []byte(secret)
		return body, "text/plain", nil
	}

	file, header, err := r.FormFile("secret")
	contentType := header.Header.Get("Content-Type")

	if err != nil {
		return nil, "", err
	}

	data, err := io.ReadAll(file)

	return data, contentType, err
}

func getContentType(r *http.Request) string {
	ct := r.Header.Get("content-type")
	if ct == "" {
		ct = "application/octet-stream"
	}

	ct, _, err := mime.ParseMediaType(ct)

	if err != nil {
		return "application/octet-stream"
	}

	return ct
}

func getContent(r *http.Request) ([]byte, string, error) {
	ct := getContentType(r)
	switch {
	case ct == "multipart/form-data":
		return parseMultiForm(r)
	default:
		data, err := io.ReadAll(r.Body)
		return data, ct, err
	}
}

func (c CreateEntryParser) calculateExpiration(expire string, defaultExpire time.Duration) (time.Duration, error) {
	exp, err := expiration.Parse(expire, defaultExpire, c.maxExpireSeconds)
	if err != nil {
		return 0, ErrInvalidExpirationDate
	}

	return exp, nil
}

func (c CreateEntryParser) getSecretExpiration(r *http.Request) (time.Duration, error) {
	var expiration string
	if err := r.ParseForm(); err != nil {
		return 0, err
	}
	expiration = r.Form.Get("expire")

	return c.calculateExpiration(expiration, time.Second*time.Duration(c.maxExpireSeconds))
}

func (c CreateEntryParser) getSecretMaxReads(r *http.Request) (int, error) {
	if err := r.ParseForm(); err != nil {
		return 0, err
	}

	reads, err := maxreads.Parse(r.Form.Get("maxReads"))
	if err != nil {
		return 0, ErrInvalidMaxRead
	}

	return reads, nil
}

func (c CreateEntryParser) Parse(r *http.Request) (*CreateEntryRequestData, error) {
	body, contentType, err := getContent(r)

	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, ErrInvalidData
	}

	expiration, err := c.getSecretExpiration(r)

	if err != nil {
		return nil, err
	}

	maxReads, err := c.getSecretMaxReads(r)

	if err != nil {
		return nil, err
	}

	return &CreateEntryRequestData{
		ContentType: contentType,
		Body:        body,
		Expiration:  expiration,
		MaxReads:    maxReads,
	}, nil

}
