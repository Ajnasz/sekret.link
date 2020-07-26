package main

type EntryStorage interface {
	Create(string, []byte) error
	Get(string) ([]byte, error)
	GetAndDelete(string) ([]byte, error)
}
