package storage

import (
	"context"
	"time"

	"github.com/Ajnasz/sekret.link/entries"
)

type Transform = func(*entries.Entry) (*entries.Entry, error)

// Reader interface to get stored entry
type Reader interface {
	// Read reads the secret and deletes from the underlying storage in one step
	Read(context.Context, string) (*entries.Entry, error)
	// Close Closes connection to data storage, like database
	// Executed on application shutdown
	Close() error
}

type ConfirmReader interface {
	// ReadConfirm reads the secret and deletes from the underlying storage in one
	// step if the confirm chan receives a true value
	ReadConfirm(context.Context, string) (*entries.Entry, chan bool, error)
	// Close Closes connection to data storage, like database
	// Executed on application shutdown
	Close() error
}

// Writer interface to store and delete entry
type Writer interface {
	// Writes the secret into the remote data storege
	Write(ctx context.Context, UUID string, entry []byte, expiration time.Duration, maxReads int) (*entries.EntryMeta, error)
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

// Verifyable an interface which extends the EntryStorage with a
// VerifyDelete method
type Verifyable interface {
	Writer
	// VerifyDelete checks if the given deleteKey belongs to the given UUID
	VerifyDelete(ctx context.Context, UUID string, deleteKey string) (bool, error)
}

type VerifyConfirmReader interface {
	Verifyable
	ConfirmReader
}
