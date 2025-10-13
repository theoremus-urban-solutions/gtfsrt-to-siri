package integration

import (
	"os"
	"testing"

	gtfsrtpb "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/tests/helpers"
	"google.golang.org/protobuf/proto"
)

// TestVM_EntryCount verifies that the number of VehicleActivity entries
// matches the number of VehiclePosition entities in the input
func TestVM_EntryCount(t *testing.T) {
	// Load test data
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")

	// Count input VehiclePositions
	vpData, err := helpers.LoadProtobufFile("testdata/gtfsrt/vehicle-positions.pbf")
	if err != nil {
		t.Fatalf("Failed to load vehicle positions: %v", err)
	}
	var vpFeed gtfsrtpb.FeedMessage
	if err := proto.Unmarshal(vpData, &vpFeed); err != nil {
		t.Fatalf("Failed to parse vehicle positions: %v", err)
	}
	inputCount := len(vpFeed.Entity)

	// Load GTFS-RT data from golden files
	tuData, err := os.ReadFile("../../pbf-input/trip-updates-golden.pbf")
	if err != nil {
		t.Fatalf("Failed to load trip updates: %v", err)
	}
	saData, err := os.ReadFile("../../pbf-input/alerts-golden.pbf")
	if err != nil {
		t.Fatalf("Failed to load alerts: %v", err)
	}
	gtfsrtData, err := gtfsrt.NewGTFSRTWrapper(tuData, vpData, saData)
	if err != nil {
		t.Fatalf("Failed to create GTFS-RT wrapper: %v", err)
	}

	// Generate VM response
	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)
	result := c.GetCompleteVehicleMonitoringResponse()

	if result == nil {
		t.Fatal("VM result should not be nil")
	}

	vm := result.Siri.ServiceDelivery.VehicleMonitoringDelivery
	if len(vm) == 0 {
		t.Fatal("Should have at least one VehicleMonitoringDelivery")
	}

	outputCount := len(vm[0].VehicleActivity)

	// Verify counts match
	if outputCount != inputCount {
		t.Errorf("Entry count mismatch: input=%d VehiclePositions, output=%d VehicleActivity entries",
			inputCount, outputCount)
	} else {
		t.Logf("✓ Entry count matches: %d VehiclePositions → %d VehicleActivity", inputCount, outputCount)
	}
}

// TestET_EntryCount verifies that the number of EstimatedVehicleJourney entries
// matches the number of TripUpdate entities in the input (minus any filtered for missing static data)
func TestET_EntryCount(t *testing.T) {
	// Load test data
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")

	// Count input TripUpdates
	tuData, err := helpers.LoadProtobufFile("testdata/gtfsrt/trip-updates.pbf")
	if err != nil {
		t.Fatalf("Failed to load trip updates: %v", err)
	}
	var tuFeed gtfsrtpb.FeedMessage
	if err := proto.Unmarshal(tuData, &tuFeed); err != nil {
		t.Fatalf("Failed to parse trip updates: %v", err)
	}
	inputCount := len(tuFeed.Entity)

	// Load GTFS-RT data from golden files
	vpData, err := os.ReadFile("../../pbf-input/vehicle-positions-golden.pbf")
	if err != nil {
		t.Fatalf("Failed to load vehicle positions: %v", err)
	}
	saData, err := os.ReadFile("../../pbf-input/alerts-golden.pbf")
	if err != nil {
		t.Fatalf("Failed to load alerts: %v", err)
	}
	gtfsrtData, err := gtfsrt.NewGTFSRTWrapper(tuData, vpData, saData)
	if err != nil {
		t.Fatalf("Failed to create GTFS-RT wrapper: %v", err)
	}

	// Generate ET response
	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)
	result := c.BuildEstimatedTimetable()

	if len(result.EstimatedJourneyVersionFrame) == 0 {
		t.Fatal("Should have at least one EstimatedJourneyVersionFrame")
	}

	outputCount := len(result.EstimatedJourneyVersionFrame[0].EstimatedVehicleJourney)

	// With synchronized feeds, counts should match exactly
	// (trips in RT should exist in static)
	if outputCount != inputCount {
		filtered := inputCount - outputCount
		t.Logf("Note: %d trips filtered (exist in GTFS-RT but not in GTFS static)", filtered)

		// Allow some filtering due to data mismatches, but warn if it's excessive
		filterRate := float64(filtered) / float64(inputCount) * 100
		if filterRate > 5.0 {
			t.Errorf("High filter rate: %.1f%% of trips filtered (%d/%d). "+
				"This may indicate a data sync issue between GTFS static and GTFS-RT feeds.",
				filterRate, filtered, inputCount)
		} else {
			t.Logf("✓ Entry count acceptable: %d TripUpdates → %d EstimatedVehicleJourney (%.1f%% filtered)",
				inputCount, outputCount, filterRate)
		}
	} else {
		t.Logf("✓ Entry count matches perfectly: %d TripUpdates → %d EstimatedVehicleJourney",
			inputCount, outputCount)
	}
}

// TestSX_EntryCount verifies that the number of PtSituationElement entries
// matches the number of Alert entities in the input
func TestSX_EntryCount(t *testing.T) {
	// Load test data
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")

	// Count input Alerts
	saData, err := helpers.LoadProtobufFile("testdata/gtfsrt/alerts.pbf")
	if err != nil {
		t.Fatalf("Failed to load service alerts: %v", err)
	}
	var saFeed gtfsrtpb.FeedMessage
	if err := proto.Unmarshal(saData, &saFeed); err != nil {
		t.Fatalf("Failed to parse service alerts: %v", err)
	}
	inputCount := len(saFeed.Entity)

	// Load GTFS-RT data from golden files
	tuData, err := os.ReadFile("../../pbf-input/trip-updates-golden.pbf")
	if err != nil {
		t.Fatalf("Failed to load trip updates: %v", err)
	}
	vpData, err := os.ReadFile("../../pbf-input/vehicle-positions-golden.pbf")
	if err != nil {
		t.Fatalf("Failed to load vehicle positions: %v", err)
	}
	gtfsrtData, err := gtfsrt.NewGTFSRTWrapper(tuData, vpData, saData)
	if err != nil {
		t.Fatalf("Failed to create GTFS-RT wrapper: %v", err)
	}

	// Generate SX response
	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)
	result := c.BuildSituationExchange()

	situations, ok := result.Situations.([]siri.PtSituationElement)
	if !ok {
		t.Fatalf("Expected []siri.PtSituationElement in Situations field, got %T", result.Situations)
	}

	outputCount := len(situations)

	// Verify counts match
	if outputCount != inputCount {
		t.Errorf("Entry count mismatch: input=%d Alerts, output=%d PtSituationElement entries",
			inputCount, outputCount)
	} else {
		t.Logf("✓ Entry count matches: %d Alerts → %d PtSituationElement", inputCount, outputCount)
	}
}
