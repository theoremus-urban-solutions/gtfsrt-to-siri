package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
)

// LoadGTFSRTFromLocal loads GTFS-RT protobuf files from local testdata using raw bytes
func LoadGTFSRTFromLocal(t *testing.T) *gtfsrt.GTFSRTWrapper {
	t.Helper()

	basePath := GetTestDataPath()
	vpPath := basePath + "/gtfsrt/vehicle-positions.pb"
	tuPath := basePath + "/gtfsrt/trip-updates.pb"
	saPath := basePath + "/gtfsrt/service-alerts.pb"

	// Read raw protobuf bytes
	vpData, err := os.ReadFile(vpPath)
	if err != nil {
		t.Fatalf("Failed to read vehicle positions: %v", err)
	}

	tuData, err := os.ReadFile(tuPath)
	if err != nil {
		t.Fatalf("Failed to read trip updates: %v", err)
	}

	saData, err := os.ReadFile(saPath)
	if err != nil {
		t.Fatalf("Failed to read service alerts: %v", err)
	}

	// Create wrapper from raw bytes
	wrapper, err := gtfsrt.NewGTFSRTWrapper(tuData, vpData, saData)
	if err != nil {
		t.Fatalf("Failed to create GTFS-RT wrapper: %v", err)
	}

	return wrapper
}

// LoadProtobufFile loads a raw protobuf file from testdata/gtfsrt
func LoadProtobufFile(filename string) ([]byte, error) {
	path := filepath.Join(GetTestDataPath(), "gtfsrt", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}
	return data, nil
}
