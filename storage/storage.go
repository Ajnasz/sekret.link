package storage

import (
	"context"
	"time"

	"github.com/Ajnasz/sekret.link/entries"
)

// Reader interface to get stored entry
type Reader interface {
	GetAndDelete(context.Context, string) (*entries.Entry, error)
	GetMeta(context.Context, string) (*entries.EntryMeta, error)
	// Get(UUID string) (*entries.Entry, error)
	// Closes connection to data storage, like database
	// Executed on application shutdown
	Close() error
}

// Writer interface to store and delete entry
type Writer interface {
	// Writes the secret into the remote data storege
	Create(ctx context.Context, UUID string, entry []byte, expiration time.Duration, maxReads int) error
	Delete(context.Context, string) error
	DeleteExpired(context.Context) error
	// Closes connection to data storage, like database
	// Executed on application shutdown
	Close() error
}

// Storage interface to storea, read and delete entries
type Storage interface {
	Reader
	Writer
}

// Cleanable Interface which enables to remove every entry from a storae
type Cleanable interface {
	Storage
	Clean()
}

// Verifyable an interface which extends the EntryStorage with a
// VerifyDelete method
type Verifyable interface {
	Storage
	// VerifyDelete checks if the given deleteKey belongs to the given UUID
	VerifyDelete(ctx context.Context, UUID string, deleteKey string) (bool, error)
}
