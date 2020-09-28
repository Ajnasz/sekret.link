package main

import (
	"fmt"
	"time"

	"github.com/Ajnasz/sekret.link/storage"
)

type EntryStorage interface {
	Close() error
	Create(string, []byte, time.Duration) error
	Get(string) (*storage.Entry, error)
	GetAndDelete(string) (*storage.Entry, error)
	GetMeta(string) (*storage.EntryMeta, error)
	Delete(string) error
	DeleteExpired() error
}

type SecretResponse struct {
	UUID     string
	Key      string
	Data     string
	Created  time.Time
	Accessed time.Time
	Expire   time.Time
}

func secretResponseFromEntryMeta(meta storage.EntryMeta) *SecretResponse {
	return &SecretResponse{
		UUID:     meta.UUID,
		Created:  meta.Created,
		Expire:   meta.Expire,
		Accessed: meta.Accessed,
	}
}

var ErrEntryExpired = fmt.Errorf("Entry expired")
var ErrEntryNotFound = fmt.Errorf("Entry not found")
