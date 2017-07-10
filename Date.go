package GOHMoneyDB

import (
	"time"
)


const (
	DbDateFormat string = "2006-01-02"
)

// formatDateString formats a date into the accepted date format for use with the DB
func formatDateString(date time.Time) string {
	return date.Format(DbDateFormat)
}

// ParseDateString parses a date string from the DB into a time.Time item and returns it with any errors that occured whilst attemping to parse the string.
func ParseDateString(dateString string) (time.Time, error) {
	return time.Parse(DbDateFormat, dateString)
}