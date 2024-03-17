package parsers

import (
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers/expiration"
)

type CreateEntryParser struct {
	maxExpireSeconds int
}

type CreateEntryRequestData struct {
	Body       []byte
	Expiration time.Duration
	MaxReads   int
}

func NewCreateEntryParser(maxExpireSeconds int) CreateEntryParser {
	return CreateEntryParser{maxExpireSeconds: maxExpireSeconds}
}

func parseMultiForm(r *http.Request) ([]byte, error) {
	err := r.ParseMultipartForm(1024 * 1024)
	if err != nil {
		return nil, err
	}

	secret := r.PostForm.Get("secret")
	if secret != "" {
		body := []byte(secret)
		return body, nil
	}

	file, _, err := r.FormFile("secret")

	if err != nil {
		return nil, err
	}

	return io.ReadAll(file)
}

func getBody(r *http.Request) ([]byte, error) {
	ct := r.Header.Get("content-type")
	if ct == "" {
		ct = "application/octet-stream"
	}

	ct, _, err := mime.ParseMediaType(ct)

	if err != nil {
		return nil, err
	}

	switch {
	case ct == "multipart/form-data":
		return parseMultiForm(r)
	default:
		return io.ReadAll(r.Body)
	}
}

func (c CreateEntryParser) calculateExpiration(expire string, defaultExpire time.Duration) (time.Duration, error) {
	exp, err := expiration.CalculateExpiration(expire, defaultExpire, c.maxExpireSeconds)
	if err != nil {
		return 0, ErrInvalidExpirationDate
	}

	return exp, nil
}

func (c CreateEntryParser) getSecretExpiration(r *http.Request) (time.Duration, error) {
	var expiration string
	r.ParseForm()
	expiration = r.Form.Get("expire")

	return c.calculateExpiration(expiration, time.Second*time.Duration(c.maxExpireSeconds))
}

func getSecretMaxReads(r *http.Request) (int, error) {
	r.ParseForm()
	const minMaxReadCount int = 1
	val := r.Form.Get("maxReads")
	if val == "" {
		return minMaxReadCount, nil
	}

	maxReads, err := strconv.Atoi(val)
	if err != nil {
		if _, isNumError := err.(*strconv.NumError); isNumError {
			return 0, ErrInvalidMaxRead
		}

		return 0, err
	}

	if maxReads < minMaxReadCount {
		return 0, ErrInvalidMaxRead
	}

	return maxReads, nil
}

func (c CreateEntryParser) Parse(r *http.Request) (*CreateEntryRequestData, error) {
	body, err := getBody(r)

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

	maxReads, err := getSecretMaxReads(r)

	if err != nil {
		return nil, err
	}

	return &CreateEntryRequestData{
		Body:       body,
		Expiration: expiration,
		MaxReads:   maxReads,
	}, nil

}
