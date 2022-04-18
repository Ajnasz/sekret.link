package entries

import (
	"fmt"
)

// ErrEntryExpired returned when an expired entry requested
var ErrEntryExpired = fmt.Errorf("Entry expired")

// ErrEntryNotFound returned when an entry not found
var ErrEntryNotFound = fmt.Errorf("Entry not found")
