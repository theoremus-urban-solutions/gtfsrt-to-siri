package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"mta/gtfsrt-to-siri/config"
	"mta/gtfsrt-to-siri/gtfs"
)

// GetTestDataPath returns absolute path to testdata/
func GetTestDataPath() string {
	wd, _ := os.Getwd()
	for {
		testdataPath := filepath.Join(wd, "testdata")
		if _, err := os.Stat(testdataPath); err == nil {
			return testdataPath
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			panic("Could not find testdata directory")
		}
		wd = parent
	}
}

// LoadTestGTFS loads a GTFS fixture from testdata/gtfs/
func LoadTestGTFS(t *testing.T, filename string) *gtfs.GTFSIndex {
	t.Helper()
	path := filepath.Join(GetTestDataPath(), "gtfs", filename)

	cfg := config.GTFSConfig{
		AgencyID:  "SOFIA",
		StaticURL: path,
	}

	idx, err := gtfs.NewGTFSIndexFromConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to load GTFS fixture %s: %v", filename, err)
	}
	return idx
}

// MustLoadTestGTFS loads a GTFS fixture or panics (for init)
func MustLoadTestGTFS(filename string) *gtfs.GTFSIndex {
	path := filepath.Join(GetTestDataPath(), "gtfs", filename)

	cfg := config.GTFSConfig{
		AgencyID:  "SOFIA",
		StaticURL: path,
	}

	idx, err := gtfs.NewGTFSIndexFromConfig(cfg)
	if err != nil {
		panic("Failed to load GTFS fixture: " + err.Error())
	}
	return idx
}

// LoadTestConfig returns test configuration
func LoadTestConfig(t *testing.T) *config.AppConfig {
	t.Helper()

	// Return a basic test config
	return &config.AppConfig{
		GTFS: config.GTFSConfig{
			AgencyID: "SOFIA",
		},
		Converter: config.ConverterConfig{
			CallDistanceAlongRouteNumDigits: 2,
			TripKeyStrategy:                 "raw",
		},
	}
}
