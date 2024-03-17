package maxreads

import (
	"errors"
	"strconv"
)

// ErrInvalidMaxRead request parse error happens when the user maximum read
// number is greater than the system maximum read number
var ErrInvalidMaxRead = errors.New("Invalid max read")

// Parse returns the maximum number of reads for a secret.
func Parse(val string) (int, error) {
	const minMaxReadCount int = 1
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
