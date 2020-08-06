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
