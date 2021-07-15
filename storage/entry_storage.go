package storage

import (
	"time"

	"github.com/Ajnasz/sekret.link/entries"
)

type EntryStorage interface {
	Close() error
	Create(string, []byte, time.Duration, int) error
	// Get(string) (*Entry, error)
	GetAndDelete(string) (*entries.Entry, error)
	GetMeta(string) (*entries.EntryMeta, error)
	Delete(string) error
	DeleteExpired() error
}
