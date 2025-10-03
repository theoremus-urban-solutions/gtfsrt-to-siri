package utils

import (
	"fmt"
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

// Iso8601FromGTFSTimeAndDate converts GTFS time (HH:MM:SS) and date (YYYYMMDD) to ISO8601 timestamp
// Handles GTFS times that can be > 24:00:00 for trips that span past midnight
func Iso8601FromGTFSTimeAndDate(gtfsTime string, gtfsDate string) string {
	if gtfsTime == "" || len(gtfsDate) != 8 {
		return ""
	}

	unixSeconds := ParseGTFSTimeToUnixSeconds(gtfsTime, gtfsDate)
	if unixSeconds == 0 {
		return ""
	}

	return time.Unix(unixSeconds, 0).UTC().Format(time.RFC3339)
}

// ParseGTFSTimeToUnixSeconds converts GTFS time (HH:MM:SS) and date (YYYYMMDD) to Unix timestamp
// Handles GTFS times that can be > 24:00:00 for trips that span past midnight
// Returns 0 on error
// Uses local timezone to match ET behavior (GTFS static times are in local time)
func ParseGTFSTimeToUnixSeconds(gtfsTime string, gtfsDate string) int64 {
	if gtfsTime == "" || len(gtfsDate) != 8 {
		return 0
	}

	// Parse the base date (YYYYMMDD)
	year := gtfsDate[:4]
	month := gtfsDate[4:6]
	day := gtfsDate[6:8]

	// Parse the time (HH:MM:SS)
	var h, m, s int
	if _, err := fmt.Sscanf(gtfsTime, "%d:%d:%d", &h, &m, &s); err != nil {
		return 0
	}

	// Build date string and parse in local timezone (not UTC)
	// GTFS static times are in local time, matching ET's gtfsTimeToUnixTimestamp()
	dateStr := fmt.Sprintf("%s-%s-%sT00:00:00", year, month, day)
	t, err := time.ParseInLocation("2006-01-02T15:04:05", dateStr, time.Local)
	if err != nil {
		return 0
	}

	// Add time (handles hours >= 24)
	t = t.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second)
	return t.Unix()
}

// FormatDelayAsISO8601Duration converts delay in seconds to ISO 8601 duration format
// Positive delay: PT5M30S (5 minutes 30 seconds late)
// Negative delay: -PT2M15S (2 minutes 15 seconds early)
// Zero delay: PT0S
func FormatDelayAsISO8601Duration(delaySeconds int64) string {
	if delaySeconds == 0 {
		return "PT0S"
	}

	// Handle negative (early)
	negative := delaySeconds < 0
	if negative {
		delaySeconds = -delaySeconds
	}

	hours := delaySeconds / 3600
	minutes := (delaySeconds % 3600) / 60
	seconds := delaySeconds % 60

	result := "PT"
	if hours > 0 {
		result += fmt.Sprintf("%dH", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dM", minutes)
	}
	if seconds > 0 || (hours == 0 && minutes == 0) {
		result += fmt.Sprintf("%dS", seconds)
	}

	if negative {
		result = "-" + result
	}

	return result
}
