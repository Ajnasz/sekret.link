package main

import (
	"fmt"
	"time"
)

type EntryMeta struct {
	UUID     string
	Created  time.Time
	Accessed time.Time
	Expire   time.Time
}

func (e *EntryMeta) IsExpired() bool {
	return e.Expire.Before(time.Now())
}

type Entry struct {
	EntryMeta
	Data []byte
}

type EntryStorage interface {
	Close() error
	Create(string, []byte, time.Duration) error
	Get(string) (*Entry, error)
	GetAndDelete(string) (*Entry, error)
	GetMeta(string) (*EntryMeta, error)
	Delete(string) error
}

type SecretResponse struct {
	UUID     string
	Key      string
	Data     string
	Created  time.Time
	Accessed time.Time
	Expire   time.Time
}

func secretResponseFromEntryMeta(entry *EntryMeta) *SecretResponse {
	return &SecretResponse{
		UUID:     entry.UUID,
		Created:  entry.Created,
		Expire:   entry.Expire,
		Accessed: entry.Accessed,
	}
}

var entryExpiredError = fmt.Errorf("Entry expired")
var entryNotFound = fmt.Errorf("Entry not found")
