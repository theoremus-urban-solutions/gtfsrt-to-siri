package utils

import (
	"time"
)

// Iso8601Now returns the current time in ISO8601 format
func Iso8601Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Iso8601FromUnixSeconds converts Unix timestamp to ISO8601 format
func Iso8601FromUnixSeconds(sec int64) string {
	return time.Unix(sec, 0).UTC().Format(time.RFC3339)
}

// Iso8601ExtendedFromUnixSeconds returns timestamp with nanosecond precision and timezone
func Iso8601ExtendedFromUnixSeconds(sec int64) string {
	return time.Unix(sec, 0).Format("2006-01-02T15:04:05.000000000-07:00")
}

// Iso8601DateFromUnixSeconds returns just the date portion in YYYY-MM-DD format
func Iso8601DateFromUnixSeconds(sec int64) string {
	return time.Unix(sec, 0).UTC().Format("2006-01-02")
}

// ValidUntilFrom calculates the valid until timestamp
func ValidUntilFrom(baseEpoch int64, readIntervalMS int) string {
	if baseEpoch <= 0 || readIntervalMS <= 0 {
		return ""
	}
	return time.Unix(baseEpoch+int64(readIntervalMS/1000), 0).UTC().Format(time.RFC3339)
}
