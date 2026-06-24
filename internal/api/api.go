// Package api contains the http.Handler implementations for the api endpoints
package api

import (
	"errors"
)

// ErrInvalidExpirationDate request parse error happens when the user set
// expiration date is larger than the system maximum expiration date
var ErrInvalidExpirationDate = errors.New("invalid expiration date")

// ErrInvalidMaxRead request parse error happens when the user maximum read
// number is greater than the system maximum read number
var ErrInvalidMaxRead = errors.New("invalid max read")

// ErrInvalidData request parse error happens if the post data can not be accepted
var ErrInvalidData = errors.New("invalid data")

// ErrRequestParseError request parse error happens if the post data can not be accepted
var ErrRequestParseError = errors.New("request parse error")
