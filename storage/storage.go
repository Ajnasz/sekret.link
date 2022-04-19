package storage

// VerifyStorage an interface which extends the EntryStorage with a
// VerifyDelete method
type VerifyStorage interface {
	EntryStorage
	VerifyDelete(string, string) (bool, error)
}
