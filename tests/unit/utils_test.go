package unit

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

func TestIso8601Now(t *testing.T) {
	before := time.Now().UTC().Add(-1 * time.Second)
	result := utils.Iso8601Now()
	after := time.Now().UTC().Add(1 * time.Second)

	parsed, err := time.Parse(time.RFC3339, result)
	if err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if parsed.Before(before) || parsed.After(after) {
		t.Errorf("timestamp should be between %v and %v, got %v", before, after, parsed)
	}
}

func TestIso8601FromUnixSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "epoch",
			input:    0,
			expected: "1970-01-01T00:00:00Z",
		},
		{
			name:     "specific timestamp",
			input:    1696320000, // 2023-10-03 08:00:00 UTC
			expected: "2023-10-03T08:00:00Z",
		},
		{
			name:     "negative timestamp",
			input:    -86400, // 1 day before epoch
			expected: "1969-12-31T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.Iso8601FromUnixSeconds(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIso8601DateFromUnixSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "epoch",
			input:    0,
			expected: "1970-01-01",
		},
		{
			name:     "specific date",
			input:    1696320000,
			expected: "2023-10-03",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.Iso8601DateFromUnixSeconds(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseGTFSTimeToUnixSeconds(t *testing.T) {
	baseDate := "20251003" // 2025-10-03

	tests := []struct {
		name      string
		gtfsTime  string
		gtfsDate  string
		wantHour  int
		wantMin   int
		wantError bool
	}{
		{
			name:     "normal morning time",
			gtfsTime: "08:30:45",
			gtfsDate: baseDate,
			wantHour: 8,
			wantMin:  30,
		},
		{
			name:     "midnight",
			gtfsTime: "00:00:00",
			gtfsDate: baseDate,
			wantHour: 0,
			wantMin:  0,
		},
		{
			name:     "late night",
			gtfsTime: "23:59:59",
			gtfsDate: baseDate,
			wantHour: 23,
			wantMin:  59,
		},
		{
			name:     "past midnight (25:30:00)",
			gtfsTime: "25:30:00",
			gtfsDate: baseDate,
			wantHour: 1, // Next day 01:30
			wantMin:  30,
		},
		{
			name:     "very late (26:45:30)",
			gtfsTime: "26:45:30",
			gtfsDate: baseDate,
			wantHour: 2,
			wantMin:  45,
		},
		{
			name:      "empty time",
			gtfsTime:  "",
			gtfsDate:  baseDate,
			wantError: true,
		},
		{
			name:      "invalid date",
			gtfsTime:  "08:00:00",
			gtfsDate:  "2025",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ParseGTFSTimeToUnixSeconds(tt.gtfsTime, tt.gtfsDate)

			if tt.wantError {
				if result != 0 {
					t.Errorf("expected 0 for error case, got %d", result)
				}
				return
			}

			if result == 0 {
				t.Fatal("expected non-zero result")
			}

			parsed := time.Unix(result, 0)
			if parsed.Hour() != tt.wantHour {
				t.Errorf("hour mismatch: expected %d, got %d", tt.wantHour, parsed.Hour())
			}
			if parsed.Minute() != tt.wantMin {
				t.Errorf("minute mismatch: expected %d, got %d", tt.wantMin, parsed.Minute())
			}
		})
	}
}

func TestFormatDelayAsISO8601Duration(t *testing.T) {
	tests := []struct {
		name         string
		delaySeconds int64
		expected     string
	}{
		{
			name:         "zero delay",
			delaySeconds: 0,
			expected:     "PT0S",
		},
		{
			name:         "positive seconds only",
			delaySeconds: 45,
			expected:     "PT45S",
		},
		{
			name:         "positive minutes and seconds",
			delaySeconds: 330, // 5min 30sec
			expected:     "PT5M30S",
		},
		{
			name:         "positive minutes only",
			delaySeconds: 300, // 5min
			expected:     "PT5M",
		},
		{
			name:         "positive hours, minutes, seconds",
			delaySeconds: 7545, // 2h 5m 45s
			expected:     "PT2H5M45S",
		},
		{
			name:         "positive hours only",
			delaySeconds: 7200, // 2h
			expected:     "PT2H",
		},
		{
			name:         "negative seconds",
			delaySeconds: -30,
			expected:     "-PT30S",
		},
		{
			name:         "negative minutes",
			delaySeconds: -135, // -2min 15sec
			expected:     "-PT2M15S",
		},
		{
			name:         "negative hours",
			delaySeconds: -3665, // -1h 1m 5s
			expected:     "-PT1H1M5S",
		},
		{
			name:         "one second",
			delaySeconds: 1,
			expected:     "PT1S",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatDelayAsISO8601Duration(tt.delaySeconds)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidUntilFrom(t *testing.T) {
	tests := []struct {
		name           string
		baseEpoch      int64
		readIntervalMS int
		expected       string
	}{
		{
			name:           "valid calculation",
			baseEpoch:      1696320000, // 2023-10-03 08:00:00
			readIntervalMS: 30000,      // 30 seconds
			expected:       "2023-10-03T08:00:30Z",
		},
		{
			name:           "zero base epoch",
			baseEpoch:      0,
			readIntervalMS: 30000,
			expected:       "",
		},
		{
			name:           "negative interval",
			baseEpoch:      1696320000,
			readIntervalMS: -30000,
			expected:       "",
		},
		{
			name:           "zero interval",
			baseEpoch:      1696320000,
			readIntervalMS: 0,
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ValidUntilFrom(tt.baseEpoch, tt.readIntervalMS)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPresentableDistance(t *testing.T) {
	tests := []struct {
		name                  string
		stopsFromCurStop      int
		distToCurrentStopKM   float64
		distToImmedNextStopKM float64
		expected              string
	}{
		{
			name:                  "at stop - very close",
			stopsFromCurStop:      0,
			distToCurrentStopKM:   0.01, // ~30 feet
			distToImmedNextStopKM: 0.5,
			expected:              "at stop",
		},
		{
			name:                  "approaching",
			stopsFromCurStop:      0,
			distToCurrentStopKM:   0.1, // ~320 feet
			distToImmedNextStopKM: 0.5,
			expected:              "approaching",
		},
		{
			name:                  "one stop away",
			stopsFromCurStop:      1,
			distToCurrentStopKM:   1.0,
			distToImmedNextStopKM: 0.3,
			expected:              "1 stop",
		},
		{
			name:                  "two stops away",
			stopsFromCurStop:      2,
			distToCurrentStopKM:   2.0,
			distToImmedNextStopKM: 0.3,
			expected:              "2 stops",
		},
		{
			name:                  "five stops away",
			stopsFromCurStop:      5,
			distToCurrentStopKM:   0.5, // Keep distance small to avoid miles display
			distToImmedNextStopKM: 0.1,
			expected:              "5 stops",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.PresentableDistance(
				tt.stopsFromCurStop,
				tt.distToCurrentStopKM,
				tt.distToImmedNextStopKM,
			)

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
