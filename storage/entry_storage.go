package storage

import "time"

type EntryStorage interface {
	Close() error
	Create(string, []byte, time.Duration) error
	Get(string) (*Entry, error)
	GetAndDelete(string) (*Entry, error)
	GetMeta(string) (*EntryMeta, error)
	Delete(string) error
	DeleteExpired() error
}
