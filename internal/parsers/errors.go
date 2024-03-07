package parsers

import (
	"errors"
)

// ErrInvalidData request parse error happens if the post data can not be accepted
var ErrInvalidData = errors.New("Invalid data")

// ErrInvalidMaxRead request parse error happens when the user maximum read
// number is greater than the system maximum read number
var ErrInvalidMaxRead = errors.New("Invalid max read")

// ErrInvalidExpirationDate request parse error happens when the user set
// expiration date is larger than the system maximum expiration date
var ErrInvalidExpirationDate = errors.New("Invalid expiration date")

// ErrInvalidURL is returned when the URL is invalid
var ErrInvalidURL = errors.New("invalid URL")

// ErrInvalidUUID is returned when the UUID is invalid
var ErrInvalidUUID = errors.New("invalid UUID")
