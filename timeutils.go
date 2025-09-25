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

func validUntilFrom(baseEpoch int64, readIntervalMS int) string {
	if baseEpoch <= 0 || readIntervalMS <= 0 {
		return ""
	}
	return time.Unix(baseEpoch+int64(readIntervalMS/1000), 0).UTC().Format(time.RFC3339)
}
