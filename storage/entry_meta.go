package storage

import "time"

type EntryMeta struct {
	UUID     string
	Created  time.Time
	Accessed time.Time
	Expire   time.Time
	MaxReads int32
}

func (e *EntryMeta) IsExpired() bool {
	return e.Expire.Before(time.Now())
}
