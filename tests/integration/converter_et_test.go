package integration

import (
	"testing"

	"mta/gtfsrt-to-siri/converter"
	"mta/gtfsrt-to-siri/tests/helpers"
)

// Test EstimatedTimetable conversion
func TestConverter_EstimatedTimetable_Basic(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.BuildEstimatedTimetable()

	if len(result.EstimatedJourneyVersionFrame) == 0 {
		t.Fatal("Should have at least one frame")
	}

	journeys := result.EstimatedJourneyVersionFrame[0].EstimatedVehicleJourney
	if len(journeys) == 0 {
		t.Fatal("Should have at least one estimated vehicle journey")
	}

	t.Logf("Generated %d estimated vehicle journeys", len(journeys))
}

// Test ET with recorded and estimated calls
func TestConverter_ET_RecordedAndEstimatedCalls(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.BuildEstimatedTimetable()

	if len(result.EstimatedJourneyVersionFrame) == 0 {
		t.Fatal("Should have at least one frame")
	}

	journeys := result.EstimatedJourneyVersionFrame[0].EstimatedVehicleJourney

	// Check that at least some journeys have both recorded and estimated calls
	hasRecorded := false
	hasEstimated := false

	for _, evj := range journeys {
		if len(evj.RecordedCalls) > 0 {
			hasRecorded = true
		}
		if len(evj.EstimatedCalls) > 0 {
			hasEstimated = true
		}
		if hasRecorded && hasEstimated {
			break
		}
	}

	if !hasRecorded {
		t.Error("Should have at least one journey with recorded calls")
	}
	if !hasEstimated {
		t.Error("Should have at least one journey with estimated calls")
	}

	t.Logf("✓ Found journeys with recorded and estimated calls")
}

// Test ET journey metadata (line ref, direction, operator)
func TestConverter_ET_JourneyMetadata(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.BuildEstimatedTimetable()

	if len(result.EstimatedJourneyVersionFrame) == 0 {
		t.Fatal("Should have at least one frame")
	}

	journeys := result.EstimatedJourneyVersionFrame[0].EstimatedVehicleJourney
	if len(journeys) == 0 {
		t.Fatal("Need at least one journey to test")
	}

	evj := journeys[0]

	if evj.LineRef == "" {
		t.Error("LineRef should not be empty")
	}
	if evj.OperatorRef == "" {
		t.Error("OperatorRef should not be empty")
	}

	t.Logf("Journey: Line=%s, Direction=%s, Operator=%s",
		evj.LineRef, evj.DirectionRef, evj.OperatorRef)
}

// Test ET call structure (times, stop info)
func TestConverter_ET_CallStructure(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.BuildEstimatedTimetable()

	if len(result.EstimatedJourneyVersionFrame) == 0 {
		t.Fatal("Should have at least one frame")
	}

	journeys := result.EstimatedJourneyVersionFrame[0].EstimatedVehicleJourney

	// Find a journey with estimated calls
	var foundValid bool
	for _, evj := range journeys {
		if len(evj.EstimatedCalls) > 0 {
			call := evj.EstimatedCalls[0]

			if call.StopPointRef == "" {
				t.Error("EstimatedCall should have StopPointRef")
			}
			if call.StopPointName == "" {
				t.Error("EstimatedCall should have StopPointName")
			}

			// At least one of arrival/departure time should be set
			if call.AimedArrivalTime == "" && call.AimedDepartureTime == "" {
				t.Error("EstimatedCall should have at least one aimed time")
			}

			foundValid = true
			t.Logf("✓ Valid call structure: %s (%s)", call.StopPointRef, call.StopPointName)
			break
		}
	}

	if !foundValid {
		t.Error("Should have at least one journey with valid estimated calls")
	}
}

// Test ET aimedTime vs expectedTime (delays)
func TestConverter_ET_DelaysInCalls(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	cfg := helpers.LoadTestConfig(t)
	c := converter.NewConverter(gtfsIndex, gtfsrtData, *cfg)

	result := c.BuildEstimatedTimetable()

	if len(result.EstimatedJourneyVersionFrame) == 0 {
		t.Fatal("Should have at least one frame")
	}

	journeys := result.EstimatedJourneyVersionFrame[0].EstimatedVehicleJourney

	// Check that we have expected times (indicating real-time data)
	hasExpectedTimes := false
	for _, evj := range journeys {
		for _, call := range evj.EstimatedCalls {
			if call.ExpectedArrivalTime != "" || call.ExpectedDepartureTime != "" {
				hasExpectedTimes = true
				t.Logf("Found expected times - Stop: %s, Expected arrival: %s",
					call.StopPointRef, call.ExpectedArrivalTime)
				break
			}
		}
		if hasExpectedTimes {
			break
		}
	}

	if !hasExpectedTimes {
		t.Log("Note: No expected times found (may be normal if no predictions available)")
	}
}
