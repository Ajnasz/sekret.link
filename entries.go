package main

import "time"

type Entry struct {
	UUID     string
	Data     []byte
	Created  time.Time
	Accessed time.Time
	Expire   time.Time
}

type EntryStorage interface {
	Create(string, []byte) error
	Get(string) (*Entry, error)
	GetAndDelete(string) (*Entry, error)
}

type SecretResponse struct {
	UUID     string
	Key      string
	Data     string
	Created  time.Time
	Accessed time.Time
	Expire   time.Time
}

func secretResponseFromEntry(entry *Entry) *SecretResponse {
	return &SecretResponse{
		UUID:     entry.UUID,
		Created:  entry.Created,
		Expire:   entry.Expire,
		Accessed: entry.Accessed,
	}
}
