package main

import "github.com/Ajnasz/sekret.link/storage"

type CleanableStorage interface {
	storage.EntryStorage
	Clean()
}
