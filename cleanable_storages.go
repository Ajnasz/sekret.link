package main

type CleanableStorage interface {
	EntryStorage
	Clean()
}
