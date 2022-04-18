package entries

// Entry struct represents an entry with it's data and meta data
type Entry struct {
	EntryMeta
	Data []byte
}
