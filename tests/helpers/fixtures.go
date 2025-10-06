package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfs"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
)

// TestDataPaths contains paths to test fixtures
type TestDataPaths struct {
	GTFSZip          string
	TripUpdates      string
	VehiclePositions string
	ServiceAlerts    string
}

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

// DefaultTestDataPaths returns paths to default test fixtures
func DefaultTestDataPaths() TestDataPaths {
	basePath := GetTestDataPath()
	return TestDataPaths{
		GTFSZip:          filepath.Join(basePath, "gtfs", "sofia-static.zip"),
		TripUpdates:      filepath.Join(basePath, "gtfsrt", "trip-updates.pb"),
		VehiclePositions: filepath.Join(basePath, "gtfsrt", "vehicle-positions.pb"),
		ServiceAlerts:    filepath.Join(basePath, "gtfsrt", "service-alerts.pb"),
	}
}

// LoadTestGTFS loads a GTFS fixture from testdata/gtfs/ using raw bytes
func LoadTestGTFS(t *testing.T, filename string, agencyID string) *gtfs.GTFSIndex {
	t.Helper()
	path := filepath.Join(GetTestDataPath(), "gtfs", filename)

	gtfsBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read GTFS fixture %s: %v", filename, err)
	}

	idx, err := gtfs.NewGTFSIndexFromBytes(gtfsBytes, agencyID)
	if err != nil {
		t.Fatalf("Failed to load GTFS fixture %s: %v", filename, err)
	}
	return idx
}

// MustLoadTestGTFS loads a GTFS fixture or panics (for init)
func MustLoadTestGTFS(filename string, agencyID string) *gtfs.GTFSIndex {
	path := filepath.Join(GetTestDataPath(), "gtfs", filename)

	gtfsBytes, err := os.ReadFile(path)
	if err != nil {
		panic("Failed to read GTFS fixture: " + err.Error())
	}

	idx, err := gtfs.NewGTFSIndexFromBytes(gtfsBytes, agencyID)
	if err != nil {
		panic("Failed to load GTFS fixture: " + err.Error())
	}
	return idx
}

// LoadTestGTFSRT loads GTFS-RT fixtures from testdata/gtfsrt/ using raw bytes
func LoadTestGTFSRT(t *testing.T) *gtfsrt.GTFSRTWrapper {
	t.Helper()
	paths := DefaultTestDataPaths()

	tuBytes, err := os.ReadFile(paths.TripUpdates)
	if err != nil {
		t.Fatalf("Failed to read trip updates: %v", err)
	}

	vpBytes, err := os.ReadFile(paths.VehiclePositions)
	if err != nil {
		t.Fatalf("Failed to read vehicle positions: %v", err)
	}

	saBytes, err := os.ReadFile(paths.ServiceAlerts)
	if err != nil {
		t.Fatalf("Failed to read service alerts: %v", err)
	}

	wrapper, err := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, saBytes)
	if err != nil {
		t.Fatalf("Failed to create GTFS-RT wrapper: %v", err)
	}

	return wrapper
}

// CreateTestConverter creates a converter with test data
func CreateTestConverter(t *testing.T, agencyID string, opts converter.ConverterOptions) *converter.Converter {
	t.Helper()

	gtfsIndex := LoadTestGTFS(t, "sofia-static.zip", agencyID)
	rt := LoadTestGTFSRT(t)

	return converter.NewConverter(gtfsIndex, rt, opts)
}

// DefaultConverterOptions returns common test options
func DefaultConverterOptions(agencyID string) converter.ConverterOptions {
	return converter.ConverterOptions{
		AgencyID:       agencyID,
		ReadIntervalMS: 30000,
		FieldMutators:  converter.FieldMutators{},
	}
}
