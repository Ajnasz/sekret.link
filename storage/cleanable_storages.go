package storage

// CleanableStorage Interface which enables to remove every entry from a storae
type CleanableStorage interface {
	EntryStorage
	Clean()
}
