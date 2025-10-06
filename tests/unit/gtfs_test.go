package unit

import (
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/tests/helpers"
)

func TestGTFSIndex_LoadsSofiaData(t *testing.T) {
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	if gtfsIndex == nil {
		t.Fatal("Failed to load Sofia GTFS data")
	}

	// Verify basic data is loaded by checking we can get agency name
	agencyName := gtfsIndex.GetAgencyName()
	if agencyName == "" {
		t.Error("Agency name should not be empty")
	}

	// Check that at least some trips exist
	if len(gtfsIndex.TripStopSeq) == 0 {
		t.Error("No trip sequences loaded")
	}
}

func TestGTFSIndex_GetStopName(t *testing.T) {
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	// Find a real stop ID from a trip
	var realStopID string
	for _, stops := range gtfsIndex.TripStopSeq {
		if len(stops) > 0 {
			realStopID = stops[0]
			break
		}
	}

	if realStopID == "" {
		t.Skip("No stops found in Sofia data")
	}

	tests := []struct {
		name         string
		stopID       string
		wantNonEmpty bool
	}{
		{
			name:         "valid stop exists",
			stopID:       realStopID,
			wantNonEmpty: true,
		},
		{
			name:         "non-existent stop",
			stopID:       "NONEXISTENT_12345",
			wantNonEmpty: false,
		},
		{
			name:         "empty stop ID",
			stopID:       "",
			wantNonEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gtfsIndex.GetStopName(tt.stopID)

			if tt.wantNonEmpty && result == "" {
				t.Errorf("expected non-empty stop name for %s", tt.stopID)
			}
			if !tt.wantNonEmpty && result != "" {
				t.Errorf("expected empty stop name for %s, got %s", tt.stopID, result)
			}
		})
	}
}

func TestGTFSIndex_GetRouteType_Trams(t *testing.T) {
	// CRITICAL REGRESSION TEST: Ensure trams (TM routes) have route_type 0
	// and that we handle route_type >= 0 (not > 0)
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	// Find a TM route (tram)
	var tramRouteID string
	var tramTripID string
	for tripID := range gtfsIndex.TripStopSeq {
		if strings.HasPrefix(tripID, "TM") {
			tramTripID = tripID
			tramRouteID = gtfsIndex.GetRouteIDForTrip(tripID)
			break
		}
	}

	if tramRouteID == "" {
		t.Skip("No TM (tram) routes in Sofia data")
	}

	routeType := gtfsIndex.GetRouteType(tramRouteID)
	if routeType != 0 {
		t.Errorf("TM route %s (trip %s) should have route_type 0 (tram), got %d", tramRouteID, tramTripID, routeType)
	}

	t.Logf("✓ Found tram route %s with route_type 0", tramRouteID)
}

func TestGTFSIndex_GetRouteType_Various(t *testing.T) {
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	// Test various route prefixes in Sofia
	prefixTests := []struct {
		prefix       string
		expectedType int
		description  string
	}{
		{"A", 3, "bus"},
		{"TM", 0, "tram"},
		{"TB", 11, "trolleybus"},
	}

	for _, tt := range prefixTests {
		t.Run(tt.description, func(t *testing.T) {
			// Find a trip with this prefix
			var routeID string
			for tripID := range gtfsIndex.TripStopSeq {
				if strings.HasPrefix(tripID, tt.prefix) {
					routeID = gtfsIndex.GetRouteIDForTrip(tripID)
					break
				}
			}

			if routeID == "" {
				t.Skipf("No %s routes found in Sofia data", tt.description)
			}

			result := gtfsIndex.GetRouteType(routeID)
			if result != tt.expectedType {
				t.Errorf("%s route %s has type %d, want %d", tt.description, routeID, result, tt.expectedType)
			}
		})
	}
}

func TestGTFSIndex_GetRouteShortName(t *testing.T) {
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	// Find any trip and its route
	var testRouteID string
	for tripID := range gtfsIndex.TripStopSeq {
		testRouteID = gtfsIndex.GetRouteIDForTrip(tripID)
		if testRouteID != "" {
			break
		}
	}

	if testRouteID == "" {
		t.Skip("No routes in Sofia data")
	}

	result := gtfsIndex.GetRouteShortName(testRouteID)
	if result == "" {
		t.Errorf("Route %s should have a short name", testRouteID)
	}
}

func TestGTFSIndex_GetDepartureTime(t *testing.T) {
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	// Find a trip with stop times
	var testTripID, testStopID string
	for tripID, stops := range gtfsIndex.TripStopSeq {
		if len(stops) > 0 {
			testTripID = tripID
			testStopID = stops[0]
			break
		}
	}

	if testTripID == "" {
		t.Skip("No trips with stops in Sofia data")
	}

	// Should get departure time for first stop of trip
	result := gtfsIndex.GetDepartureTime(testTripID, testStopID)
	if result == "" {
		t.Errorf("Trip %s stop %s should have departure time", testTripID, testStopID)
	}

	// Time should be in HH:MM:SS format (or HH+:MM:SS for after midnight)
	if len(result) < 8 || !strings.Contains(result, ":") {
		t.Errorf("Departure time %s doesn't look like HH:MM:SS", result)
	}

	// Non-existent trip should return empty
	result = gtfsIndex.GetDepartureTime("NONEXISTENT", testStopID)
	if result != "" {
		t.Errorf("Non-existent trip should return empty departure time")
	}
}

func TestGTFSIndex_GetArrivalTime(t *testing.T) {
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	// Find a trip with stop times
	var testTripID, testStopID string
	for tripID, stops := range gtfsIndex.TripStopSeq {
		if len(stops) > 0 {
			testTripID = tripID
			testStopID = stops[0]
			break
		}
	}

	if testTripID == "" {
		t.Skip("No trips with stops in Sofia data")
	}

	result := gtfsIndex.GetArrivalTime(testTripID, testStopID)
	if result == "" {
		t.Errorf("Trip %s stop %s should have arrival time", testTripID, testStopID)
	}
}

func TestGTFSIndex_GetPreviousStopIDOfStopForTrip(t *testing.T) {
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	// Find a trip with multiple stops
	var testTripID string
	var stops []string
	for tripID, tripStops := range gtfsIndex.TripStopSeq {
		if len(tripStops) >= 3 {
			testTripID = tripID
			stops = tripStops
			break
		}
	}

	if testTripID == "" {
		t.Skip("No trips with multiple stops in Sofia data")
	}

	// First stop should have no previous
	result := gtfsIndex.GetPreviousStopIDOfStopForTrip(testTripID, stops[0])
	if result != "" {
		t.Errorf("First stop should have no previous, got %s", result)
	}

	// Second stop's previous should be first stop
	result = gtfsIndex.GetPreviousStopIDOfStopForTrip(testTripID, stops[1])
	if result != stops[0] {
		t.Errorf("Second stop's previous should be %s, got %s", stops[0], result)
	}

	// Third stop's previous should be second stop
	result = gtfsIndex.GetPreviousStopIDOfStopForTrip(testTripID, stops[2])
	if result != stops[1] {
		t.Errorf("Third stop's previous should be %s, got %s", stops[1], result)
	}
}

func TestGTFSIndex_TripStopSequence(t *testing.T) {
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	// Find a trip
	var testTripID string
	for tripID := range gtfsIndex.TripStopSeq {
		testTripID = tripID
		break
	}

	if testTripID == "" {
		t.Skip("No trips in Sofia data")
	}

	stopSeq := gtfsIndex.TripStopSeq[testTripID]
	if len(stopSeq) == 0 {
		t.Errorf("Trip %s should have stops", testTripID)
	}

	// Verify all stops have names
	for i, stopID := range stopSeq {
		name := gtfsIndex.GetStopName(stopID)
		if name == "" {
			t.Errorf("Stop %d (%s) in trip %s has no name", i, stopID, testTripID)
		}
	}
}

func TestGTFSIndex_GetAgencyName(t *testing.T) {
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	result := gtfsIndex.GetAgencyName()
	if result == "" {
		t.Error("Agency name should not be empty")
	}

	t.Logf("Agency name: %s", result)
}

func TestGTFSIndex_GetOriginAndDestination(t *testing.T) {
	gtfsIndex := helpers.LoadTestGTFS(t, "sofia-static.zip", "SOFIA")

	// Find a trip with multiple stops
	var testTripID string
	var stops []string
	for tripID, tripStops := range gtfsIndex.TripStopSeq {
		if len(tripStops) >= 2 {
			testTripID = tripID
			stops = tripStops
			break
		}
	}

	if testTripID == "" {
		t.Skip("No trips with stops in Sofia data")
	}

	origin := gtfsIndex.GetOriginStopIDForTrip(testTripID)
	dest := gtfsIndex.GetDestinationStopIDForTrip(testTripID)

	if origin == "" {
		t.Error("Origin should not be empty")
	}
	if dest == "" {
		t.Error("Destination should not be empty")
	}

	// Origin should be first stop in sequence
	if origin != stops[0] {
		t.Errorf("Origin %s doesn't match first stop %s", origin, stops[0])
	}

	// Destination should be last stop in sequence
	if dest != stops[len(stops)-1] {
		t.Errorf("Destination %s doesn't match last stop %s", dest, stops[len(stops)-1])
	}
}
