package entries

import "time"

// EntryMeta represents the meta information of an entry, without the actual data
type EntryMeta struct {
	UUID      string
	Created   time.Time
	Accessed  time.Time
	Expire    time.Time
	DeleteKey string
	MaxReads  int32
}

// IsExpired returns true if the entry is already expired
func (e *EntryMeta) IsExpired() bool {
	return e.Expire.Before(time.Now())
}
