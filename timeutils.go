package gtfsrtsiri

import (
	"time"
)

func iso8601Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func iso8601FromUnixSeconds(sec int64) string {
	return time.Unix(sec, 0).UTC().Format(time.RFC3339)
}

// iso8601ExtendedFromUnixSeconds returns timestamp with nanosecond precision and timezone
func iso8601ExtendedFromUnixSeconds(sec int64) string {
	return time.Unix(sec, 0).Format("2006-01-02T15:04:05.000000000-07:00")
}

// iso8601DateFromUnixSeconds returns just the date portion in YYYY-MM-DD format
func iso8601DateFromUnixSeconds(sec int64) string {
	return time.Unix(sec, 0).UTC().Format("2006-01-02")
}

func validUntilFrom(baseEpoch int64, readIntervalMS int) string {
	if baseEpoch <= 0 || readIntervalMS <= 0 {
		return ""
	}
	return time.Unix(baseEpoch+int64(readIntervalMS/1000), 0).UTC().Format(time.RFC3339)
}
