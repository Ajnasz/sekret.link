package maxreads

import (
	"errors"
	"strconv"
)

// ErrInvalidMaxRead request parse error happens when the user maximum read
// number is greater than the system maximum read number
var ErrInvalidMaxRead = errors.New("invalid max read")

// Parse returns the maximum number of reads for a secret.
func Parse(val string) (int, error) {
	const minMaxReadCount int = 1
	// remaining_reads is stored in a SMALLINT column (max 32767); anything
	// above that would fail the INSERT and surface as a 500.
	const maxMaxReadCount int = 32767
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

	if maxReads < minMaxReadCount || maxReads > maxMaxReadCount {
		return 0, ErrInvalidMaxRead
	}

	return maxReads, nil
}
