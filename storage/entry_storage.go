package storage

import (
	"time"

	"github.com/Ajnasz/sekret.link/entries"
)

// EntryStorageReader interface to get stored entry
type EntryStorageReader interface {
	GetAndDelete(string) (*entries.Entry, error)
	GetMeta(string) (*entries.EntryMeta, error)
	// Closes connection to data storage, like database
	// Executed on application shutdown
	Close() error
}

// EntryStorageWriter interface to store and delete entry
type EntryStorageWriter interface {
	// Writes the secret into the remote data storege
	Create(UUID string, entry []byte, expiration time.Duration, maxReads int) error
	Delete(string) error
	DeleteExpired() error
	// Closes connection to data storage, like database
	// Executed on application shutdown
	Close() error
}

// EntryStorage interface to storea, read and delete entries
type EntryStorage interface {
	EntryStorageReader
	EntryStorageWriter
}
