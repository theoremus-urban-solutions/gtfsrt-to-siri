package helpers

import (
	"os"
	"testing"

	"mta/gtfsrt-to-siri/gtfsrt"

	gtfsrtpb "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

// LoadGTFSRTFromLocal loads GTFS-RT protobuf files from local testdata
func LoadGTFSRTFromLocal(t *testing.T) *gtfsrt.GTFSRTWrapper {
	t.Helper()

	basePath := GetTestDataPath()
	vpPath := basePath + "/gtfsrt/vehicle-positions.pb"
	tuPath := basePath + "/gtfsrt/trip-updates.pb"
	saPath := basePath + "/gtfsrt/service-alerts.pb"

	// Read protobuf files
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

	// Parse protobuf messages
	var vpFeed gtfsrtpb.FeedMessage
	if err := proto.Unmarshal(vpData, &vpFeed); err != nil {
		t.Fatalf("Failed to parse vehicle positions: %v", err)
	}

	var tuFeed gtfsrtpb.FeedMessage
	if err := proto.Unmarshal(tuData, &tuFeed); err != nil {
		t.Fatalf("Failed to parse trip updates: %v", err)
	}

	var saFeed gtfsrtpb.FeedMessage
	if err := proto.Unmarshal(saData, &saFeed); err != nil {
		t.Fatalf("Failed to parse service alerts: %v", err)
	}

	// Create wrapper and populate data using RefreshFromFeeds
	wrapper := gtfsrt.NewGTFSRTWrapper("", "", "")
	if err := wrapper.RefreshFromFeeds(&vpFeed, &tuFeed, &saFeed); err != nil {
		t.Fatalf("Failed to populate GTFS-RT wrapper: %v", err)
	}

	return wrapper
}
