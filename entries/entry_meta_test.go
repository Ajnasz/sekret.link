package entries

import (
	"testing"
	"time"
)

func Test_EntryMeta(t *testing.T) {
	expire := time.Now()

	meta := EntryMeta{Expire: expire.Add(time.Second)}

	if meta.IsExpired() {
		t.Error("entry meta should not be expired")
	}

	meta = EntryMeta{Expire: expire.Add(-time.Second)}
	if !meta.IsExpired() {
		t.Error("entry meta should be expired")
	}
}
