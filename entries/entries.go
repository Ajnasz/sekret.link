package entries

import (
	"fmt"
)

// ErrEntryExpired returned when an expired entry requested
var ErrEntryExpired = fmt.Errorf("Entry expired")
var ErrEntryNotFound = fmt.Errorf("Entry not found")
