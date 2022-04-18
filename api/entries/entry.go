package entries

import (
	"time"

	"github.com/Ajnasz/sekret.link/entries"
)

// SecretResponse the http response representation of a secret
type SecretResponse struct {
	UUID      string
	Key       string
	Data      string
	Created   time.Time
	Accessed  time.Time
	Expire    time.Time
	DeleteKey string
}

// SecretResponseFromEntryMeta SecretResponse with only meta data
func SecretResponseFromEntryMeta(meta entries.EntryMeta) SecretResponse {
	return SecretResponse{
		UUID:      meta.UUID,
		Created:   meta.Created,
		Expire:    meta.Expire,
		Accessed:  meta.Accessed,
		DeleteKey: meta.DeleteKey,
	}
}
