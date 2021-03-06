package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"
)

type requestData struct {
	body       []byte
	expiration time.Duration
	maxReads   int
}

func (c CreateHandler) calculateExpiration(expire string, defaultExpire time.Duration) (time.Duration, error) {
	if expire == "" {
		return defaultExpire, nil
	}

	userExpire, err := time.ParseDuration(expire)
	if err != nil {
		return 0, err
	}

	maxExpire := time.Duration(c.config.MaxExpireSeconds) * time.Second

	if userExpire > maxExpire {
		return 0, fmt.Errorf("Invalid expiration date")
	}

	return userExpire, nil
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

func (b requestData) ContentType() string {
	return "plain/text"
}

const MIN_MAX_READ_COUNT int = 1

func getSecretMaxReads(r *http.Request) (int, error) {
	r.ParseForm()
	val := r.Form.Get("maxReads")
	if val == "" {
		return MIN_MAX_READ_COUNT, nil
	}

	maxReads, err := strconv.Atoi(val)
	if err != nil {
		if _, isNumError := err.(*strconv.NumError); isNumError {
			return 0, fmt.Errorf("Invalid maxReads")
		}

		return 0, err
	}

	if maxReads < MIN_MAX_READ_COUNT {
		return 0, fmt.Errorf("Invalid maxReads")
	}

	return maxReads, nil
}

func (c CreateHandler) getSecretExpiration(r *http.Request) (time.Duration, error) {
	var expiration string
	r.ParseForm()
	expiration = r.Form.Get("expire")

	return c.calculateExpiration(expiration, time.Second*time.Duration(c.config.ExpireSeconds))
}

func (c CreateHandler) parseCreateRequest(r *http.Request) (*requestData, error) {
	body, err := getBody(r)

	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("Invalid data")
	}

	expiration, err := c.getSecretExpiration(r)

	if err != nil {
		return nil, err
	}

	maxReads, err := getSecretMaxReads(r)

	if err != nil {
		return nil, err
	}

	return &requestData{
		body:       body,
		expiration: expiration,
		maxReads:   maxReads,
	}, nil

}
