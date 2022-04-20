package storage

import (
	"context"
	"time"

	"github.com/Ajnasz/sekret.link/entries"
)

// EntryStorageReader interface to get stored entry
type EntryStorageReader interface {
	GetAndDelete(context.Context, string) (*entries.Entry, error)
	GetMeta(context.Context, string) (*entries.EntryMeta, error)
	// Get(UUID string) (*entries.Entry, error)
	// Closes connection to data storage, like database
	// Executed on application shutdown
	Close() error
}

// EntryStorageWriter interface to store and delete entry
type EntryStorageWriter interface {
	// Writes the secret into the remote data storege
	Create(ctx context.Context, UUID string, entry []byte, expiration time.Duration, maxReads int) error
	Delete(context.Context, string) error
	DeleteExpired(context.Context) error
	// Closes connection to data storage, like database
	// Executed on application shutdown
	Close() error
}

// EntryStorage interface to storea, read and delete entries
type EntryStorage interface {
	EntryStorageReader
	EntryStorageWriter
}

// CleanableStorage Interface which enables to remove every entry from a storae
type CleanableStorage interface {
	EntryStorage
	Clean()
}

// VerifyStorage an interface which extends the EntryStorage with a
// VerifyDelete method
type VerifyStorage interface {
	EntryStorage
	VerifyDelete(context.Context, string, string) (bool, error)
}
