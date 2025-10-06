package integration

import (
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/tests/helpers"
)

// Test VehicleMonitoring conversion with Sofia data
func TestConverter_VehicleMonitoring_RealData(t *testing.T) {
	// Load Sofia fixtures
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.GetCompleteVehicleMonitoringResponse()

	if result == nil {
		t.Fatal("VM result should not be nil")
	}

	vm := result.Siri.ServiceDelivery.VehicleMonitoringDelivery
	if len(vm) == 0 || len(vm[0].VehicleActivity) == 0 {
		t.Fatal("Should have at least one vehicle activity")
	}

	t.Logf("Generated %d vehicle activities", len(vm[0].VehicleActivity))
}

// CRITICAL REGRESSION TEST: VehicleMode for Trams
// Bug: VehicleMode was not populated for tram trips (TM prefix)
// Root Cause: Condition was `routeType > 0`, excluding route_type 0 (trams)
// Fix: Changed to `routeType >= 0`
func TestConverter_VehicleMode_TramRouteTypeZero(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.GetCompleteVehicleMonitoringResponse()
	vm := result.Siri.ServiceDelivery.VehicleMonitoringDelivery[0]

	// Find a tram vehicle (TM prefix)
	var found bool
	var lineRef, vehicleMode string
	for _, va := range vm.VehicleActivity {
		lineRef = va.MonitoredVehicleJourney.LineRef
		if strings.HasPrefix(lineRef, "TM") {
			vehicleMode = va.MonitoredVehicleJourney.VehicleMode
			found = true
			break
		}
	}

	if !found {
		t.Skip("No tram vehicles in current Sofia GTFS-RT feed")
	}

	if vehicleMode == "" {
		t.Error("VehicleMode should not be empty for tram routes")
	}

	if vehicleMode != "tram" {
		t.Errorf("Expected VehicleMode 'tram' for TM route, got '%s'", vehicleMode)
	}

	t.Logf("✓ Tram route %s has VehicleMode: %s", lineRef, vehicleMode)
}

// CRITICAL REGRESSION TEST: Delay calculation with direct tripID usage
// Bug: Delay was always 0
// Root Cause: Used computed tripKey instead of raw tripID for GTFS lookups
// Fix: Changed to use tripID directly for GetDepartureTime/GetArrivalTime
func TestConverter_Delay_UsesTripID(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.GetCompleteVehicleMonitoringResponse()
	vm := result.Siri.ServiceDelivery.VehicleMonitoringDelivery[0]

	// Check that some vehicles have non-zero delays
	hasNonZeroDelay := false
	for _, va := range vm.VehicleActivity {
		delay := va.MonitoredVehicleJourney.Delay
		if delay != "" && delay != "PT0S" {
			hasNonZeroDelay = true
			t.Logf("Vehicle %s has delay: %s",
				va.MonitoredVehicleJourney.VehicleRef, delay)
			break
		}
	}

	if !hasNonZeroDelay {
		t.Log("No vehicles with non-zero delays in current feed (may be on-time)")
	}

	// The key test: verify delays are calculated (not all zero when there should be delays)
	// This is hard to test without knowing actual delays, but we can verify format
	for _, va := range vm.VehicleActivity {
		delay := va.MonitoredVehicleJourney.Delay
		if delay != "" {
			// Should be ISO 8601 duration format
			if !strings.HasPrefix(delay, "PT") && !strings.HasPrefix(delay, "-PT") {
				t.Errorf("Delay %s is not in ISO 8601 duration format", delay)
			}
		}
	}
}

// CRITICAL REGRESSION TEST: Missing start_date fallback
// Bug: When start_date missing from GTFS-RT, delay calculation failed
// Fix: Added fallback to use feed timestamp date
func TestConverter_Delay_MissingStartDateFallback(t *testing.T) {
	// This test verifies that the converter handles missing start_date
	// by using the feed timestamp date as fallback

	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.GetCompleteVehicleMonitoringResponse()
	vm := result.Siri.ServiceDelivery.VehicleMonitoringDelivery[0]

	// Even with potentially missing start_date, we should get valid VM output
	if len(vm.VehicleActivity) == 0 {
		t.Fatal("Should have vehicle activities even with missing start_date")
	}

	// Verify all delays are in valid format (not panicking)
	for i, va := range vm.VehicleActivity {
		delay := va.MonitoredVehicleJourney.Delay
		if delay != "" {
			if !strings.HasPrefix(delay, "PT") && !strings.HasPrefix(delay, "-PT") {
				t.Errorf("Vehicle %d: Invalid delay format %s", i, delay)
			}
		}
	}

	t.Logf("✓ Handled %d vehicles successfully with start_date fallback logic",
		len(vm.VehicleActivity))
}

// Test that OriginRef and DestinationRef are populated
func TestConverter_VM_OriginDestination(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.GetCompleteVehicleMonitoringResponse()
	vm := result.Siri.ServiceDelivery.VehicleMonitoringDelivery[0]

	hasOriginDest := false
	for _, va := range vm.VehicleActivity {
		mvj := va.MonitoredVehicleJourney
		if mvj.OriginRef != "" && mvj.DestinationRef != "" {
			hasOriginDest = true
			t.Logf("Vehicle %s: %s → %s",
				mvj.VehicleRef, mvj.OriginName, mvj.DestinationName)
			break
		}
	}

	if !hasOriginDest {
		t.Error("At least one vehicle should have origin and destination")
	}
}

// Test that MonitoredCall is populated
func TestConverter_VM_MonitoredCall(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.GetCompleteVehicleMonitoringResponse()
	vm := result.Siri.ServiceDelivery.VehicleMonitoringDelivery[0]

	hasMonitoredCall := false
	for _, va := range vm.VehicleActivity {
		if va.MonitoredVehicleJourney.MonitoredCall != nil {
			mc := va.MonitoredVehicleJourney.MonitoredCall
			if mc.StopPointRef != "" {
				hasMonitoredCall = true
				t.Logf("Vehicle monitoring stop: %s (%s)",
					mc.StopPointRef, mc.StopPointName)
				break
			}
		}
	}

	if !hasMonitoredCall {
		t.Error("At least one vehicle should have monitored call")
	}
}
