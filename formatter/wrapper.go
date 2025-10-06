package formatter

import (
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

// BuildServiceDelivery creates a standardized ServiceDelivery wrapper
// with ResponseTimestamp and ProducerRef (codespace)
func BuildServiceDelivery(timestamp int64, codespace string) siri.VehicleAndSituation {
	if codespace == "" {
		codespace = "UNKNOWN"
	}

	return siri.VehicleAndSituation{
		ResponseTimestamp: utils.Iso8601FromUnixSeconds(timestamp),
		ProducerRef:       codespace,
	}
}

// WrapEstimatedTimetableResponse wraps an ET delivery in a complete SIRI response
func WrapEstimatedTimetableResponse(et siri.EstimatedTimetable, codespace string) *siri.SiriResponse {
	// Extract timestamp from ET's ResponseTimestamp
	timestamp := extractTimestampFromISO8601(et.ResponseTimestamp)

	sd := BuildServiceDelivery(timestamp, codespace)
	sd.EstimatedTimetableDelivery = []siri.EstimatedTimetable{et}

	return &siri.SiriResponse{
		Siri: siri.SiriServiceDelivery{
			ServiceDelivery: sd,
		},
	}
}

// WrapSituationExchangeResponse wraps a SX delivery in a complete SIRI response
func WrapSituationExchangeResponse(sx siri.SituationExchange, timestamp int64, codespace string) *siri.SiriResponse {
	sd := BuildServiceDelivery(timestamp, codespace)
	sd.SituationExchangeDelivery = []siri.SituationExchange{sx}

	return &siri.SiriResponse{
		Siri: siri.SiriServiceDelivery{
			ServiceDelivery: sd,
		},
	}
}

// FilterEstimatedTimetable applies filters to ET journeys
func FilterEstimatedTimetable(et siri.EstimatedTimetable, monitoringRef, lineRef, directionRef string) siri.EstimatedTimetable {
	monitoringRef = strings.ToLower(strings.TrimSpace(monitoringRef))
	lineRef = strings.ToLower(strings.TrimSpace(lineRef))
	directionRef = strings.ToLower(strings.TrimSpace(directionRef))

	filtered := siri.EstimatedTimetable{
		ResponseTimestamp:            et.ResponseTimestamp,
		EstimatedJourneyVersionFrame: []siri.EstimatedJourneyVersionFrame{},
	}

	for _, frame := range et.EstimatedJourneyVersionFrame {
		filteredJourneys := []siri.EstimatedVehicleJourney{}

		for _, journey := range frame.EstimatedVehicleJourney {
			// Filter by LineRef
			if lineRef != "" && !strings.Contains(strings.ToLower(journey.LineRef), lineRef) {
				continue
			}

			// Filter by DirectionRef
			if directionRef != "" && strings.ToLower(journey.DirectionRef) != directionRef {
				continue
			}

			// Filter by MonitoringRef (stop)
			if monitoringRef != "" {
				hasStop := false
				for _, call := range journey.RecordedCalls {
					if strings.Contains(strings.ToLower(call.StopPointRef), monitoringRef) {
						hasStop = true
						break
					}
				}
				if !hasStop {
					for _, call := range journey.EstimatedCalls {
						if strings.Contains(strings.ToLower(call.StopPointRef), monitoringRef) {
							hasStop = true
							break
						}
					}
				}
				if !hasStop {
					continue
				}
			}

			filteredJourneys = append(filteredJourneys, journey)
		}

		if len(filteredJourneys) > 0 {
			filteredFrame := siri.EstimatedJourneyVersionFrame{
				RecordedAtTime:          frame.RecordedAtTime,
				EstimatedVehicleJourney: filteredJourneys,
			}
			filtered.EstimatedJourneyVersionFrame = append(filtered.EstimatedJourneyVersionFrame, filteredFrame)
		}
	}

	return filtered
}

// extractTimestampFromISO8601 attempts to parse ISO8601 timestamp back to Unix epoch
// If parsing fails, returns current time
func extractTimestampFromISO8601(iso string) int64 {
	if iso == "" {
		return time.Now().Unix()
	}
	// Try parsing common ISO8601 formats
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000000000Z07:00",
		"2006-01-02T15:04:05Z07:00",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, iso); err == nil {
			return t.Unix()
		}
	}
	// Fallback to current time if parsing fails
	return time.Now().Unix()
}
