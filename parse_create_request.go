package main

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"time"
)

type requestData struct {
	body       []byte
	expiration time.Duration
}

func calculateExpiration(expire string, defaultExpire time.Duration) (time.Duration, error) {
	if expire == "" {
		return defaultExpire, nil
	}

	userExpire, err := time.ParseDuration(expire)
	if err != nil {
		return 0, err
	}

	maxExpire := time.Duration(maxExpireSeconds) * time.Second

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

	return ioutil.ReadAll(file)
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
		return ioutil.ReadAll(r.Body)
	}
}

func (b requestData) ContentType() string {
	return "plain/text"
}

func getSecretExpiration(r *http.Request) (time.Duration, error) {
	var expiration string
	r.ParseForm()
	expiration = r.Form.Get("expire")

	return calculateExpiration(expiration, time.Second*time.Duration(expireSeconds))
}

func parseCreateRequest(r *http.Request) (*requestData, error) {
	body, err := getBody(r)

	if err != nil {
		return nil, err
	}

	expiration, err := getSecretExpiration(r)

	if err != nil {
		return nil, err
	}

	return &requestData{
		body:       body,
		expiration: expiration,
	}, nil

}
