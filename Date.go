package GOHMoneyDB

import (
	"time"
)

// parseDateString parses a date string from the DB into a time.Time item and returns it with any errors that occured whilst attemping to parse the string.
func parseDateString(dateString string) (time.Time, error) {
	return time.Parse("2006-01-02", dateString)
}