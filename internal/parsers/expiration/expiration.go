package expiration

import (
	"errors"
	"time"
)

// ErrInvalidExpirationDate request parse error happens when the user set
// expiration date is larger than the system maximum expiration date
var ErrInvalidExpirationDate = errors.New("Invalid expiration date")

func CalculateExpiration(expire string, defaultExpire time.Duration, maxExpireSeconds int) (time.Duration, error) {
	if expire == "" {
		return defaultExpire, nil
	}

	userExpire, err := time.ParseDuration(expire)
	if err != nil {
		return 0, err
	}

	maxExpire := time.Duration(maxExpireSeconds) * time.Second

	if userExpire > maxExpire {
		return 0, ErrInvalidExpirationDate
	}

	if userExpire <= 0 {
		return 0, ErrInvalidExpirationDate
	}

	return userExpire, nil
}
