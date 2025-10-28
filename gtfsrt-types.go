// GTFS-RT Types Compilation
// This file documents all GTFS-RT types and interfaces used in this project
// for potential extraction to an external library (similar to transit-types/siri)

package main

// ==============================================================================
// CUSTOM TYPES (defined in gtfsrt/types.go)
// ==============================================================================

// RTAlert is a simplified representation of a GTFS-RT Alert for SX building
// This is our custom wrapper around the protobuf Alert message
type RTAlert struct {
	ID                string
	Header            string
	Description       string
	DescriptionByLang map[string]string // language -> description text
	URLByLang         map[string]string // language -> URL
	Cause             string
	Effect            string
	Severity          string
	Start             int64
	End               int64
	RouteIDs          []string
	StopIDs           []string
	TripIDs           []string
}

// ==============================================================================
// WRAPPER INTERFACE (what the converter expects)
// ==============================================================================

// This is the implicit interface that GTFSRTWrapper implements
// The converter depends on these methods
type GTFSRTDataSource interface {
	// Trip accessors
	GetAllMonitoredTrips() []string
	GetTripsFromTripUpdates() []string
	GetTripsFromVehiclePositions() []string
	GetGTFSTripKeyForRealtimeTripKey(tripID string) string

	// Trip metadata
	GetRouteIDForTrip(tripID string) string
	GetRouteDirectionForTrip(tripID string) string
	GetStartDateForTrip(tripID string) string
	GetOriginTimeForTrip(tripID string) string

	// Stop sequence and timing
	GetOnwardStopIDsForTrip(tripID string) []string
	GetCurrentStopIDForTrip(tripID string) string
	GetExpectedArrivalTimeAtStopForTrip(tripID, stopID string) int64
	GetExpectedDepartureTimeAtStopForTrip(tripID, stopID string) int64
	GetIndexOfStopInStopTimeUpdatesForTrip(tripID, stopID string) int
	GetScheduleRelationshipForStop(tripID, stopID string) int32

	// Timestamps
	GetVehiclePositionTimestamp(tripID string) int64
	GetTimestampForTrip(tripID string) int64
	GetTimestampForFeedMessage() int64

	// Vehicle position data
	GetVehicleRefForTrip(tripID string) string
	GetVehicleLatForTrip(tripID string) (float64, bool)
	GetVehicleLonForTrip(tripID string) (float64, bool)
	GetVehicleBearingForTrip(tripID string) (float64, bool)
	GetVehicleSpeedForTrip(tripID string) (float64, bool)

	// Occupancy and congestion
	GetOccupancyStatusForTrip(tripID string) int32
	GetCongestionLevelForTrip(tripID string) int32

	// Alerts (for SX/Situation Exchange)
	GetAlerts() []RTAlert
	GetAlertIndicesByRoute(routeID string) []int
	GetAlertIndicesByStop(stopID string) []int
	GetAlertIndicesByTrip(tripID string) []int

	// Legacy alert methods (currently unused)
	GetAllTripsWithAlert() []string
	GetTrainsWithAlertFilterObject(trips []string) map[string]bool
	GetStopsWithAlertFilterObject(trips []string) map[string]bool
	GetRoutesWithAlertFilterObject(trips []string) map[string]bool
}

// ==============================================================================
// PROTOBUF DEPENDENCIES
// ==============================================================================

// We currently use: github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs
// This provides the official GTFS-RT protobuf definitions:
//
// Key types used:
// - gtfsrtpb.FeedMessage
// - gtfsrtpb.FeedEntity
// - gtfsrtpb.TripUpdate
// - gtfsrtpb.VehiclePosition
// - gtfsrtpb.Alert
// - gtfsrtpb.TripDescriptor
// - gtfsrtpb.VehicleDescriptor
// - gtfsrtpb.StopTimeUpdate
// - gtfsrtpb.TranslatedString
//
// Enums used:
// - gtfsrtpb.TripDescriptor_ScheduleRelationship (SCHEDULED=0, SKIPPED=1, NO_DATA=2)
// - gtfsrtpb.VehiclePosition_OccupancyStatus (EMPTY=0, MANY_SEATS_AVAILABLE=1, etc.)
// - gtfsrtpb.VehiclePosition_CongestionLevel (UNKNOWN_CONGESTION_LEVEL=0, etc.)
// - gtfsrtpb.Alert_Cause
// - gtfsrtpb.Alert_Effect
// - gtfsrtpb.Alert_SeverityLevel

// ==============================================================================
// IMPLEMENTATION NOTES
// ==============================================================================

// GTFSRTWrapper (in gtfsrt/wrapper.go) is our in-memory data structure that:
// 1. Parses raw protobuf bytes from TripUpdates, VehiclePositions, and ServiceAlerts
// 2. Indexes the data for fast lookups by trip_id, route_id, stop_id
// 3. Provides the GTFSRTDataSource interface to the converter
//
// Key design decisions:
// - Data-source agnostic: accepts raw bytes, doesn't handle HTTP/file I/O
// - Separate tracking of trips from TU vs VP (for ET vs VM generation)
// - Maps for O(1) lookups instead of scanning arrays
// - Custom RTAlert type to simplify alert handling vs raw protobuf

// ==============================================================================
// POTENTIAL EXTRACTION STRATEGY
// ==============================================================================

// Option 1: Extract just the interface and RTAlert
// - Create transit-types/gtfsrt package with GTFSRTDataSource interface
// - Keep GTFSRTWrapper implementation here (it's app-specific)
// - Pro: Minimal change, clear contract
// - Con: Still depends on MobilityData protobuf bindings

// Option 2: Extract interface + wrapper + helper utilities
// - Move entire gtfsrt package to transit-types
// - Include parsing logic, helper functions
// - Pro: Reusable across projects
// - Con: More complex, need to handle protobuf dependency

// Option 3: Create abstraction layer over protobuf types
// - Define our own GTFS-RT types (not protobuf-dependent)
// - Provide adapters for MobilityData bindings
// - Pro: Clean separation, no protobuf in public API
// - Con: Most work, potential performance overhead

// ==============================================================================
// USAGE EXAMPLE
// ==============================================================================

/*
// Current usage pattern:
import (
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
	gtfsrtpb "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

// 1. Fetch raw protobuf data (from HTTP, Kafka, files, etc.)
tuBytes := fetchTripUpdates()
vpBytes := fetchVehiclePositions()
saBytes := fetchServiceAlerts()

// 2. Create wrapper
wrapper, err := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, saBytes)

// 3. Use in converter
converter := converter.NewConverter(gtfsIndex, wrapper, opts)
vm := converter.GetCompleteVehicleMonitoringResponse()
et := converter.BuildEstimatedTimetable()
sx := converter.BuildSituationExchange()
*/

// ==============================================================================
// CONSTANTS AND ENUMS
// ==============================================================================

// GTFS-RT Schedule Relationship values (from protobuf)
const (
	ScheduleRelationshipScheduled = 0
	ScheduleRelationshipSkipped   = 1
	ScheduleRelationshipNoData    = 2
)

// GTFS-RT Occupancy Status values (from protobuf)
const (
	OccupancyEmpty                   = 0
	OccupancyManySeatsAvailable      = 1
	OccupancyFewSeatsAvailable       = 2
	OccupancyStandingRoomOnly        = 3
	OccupancyCrushedStandingRoomOnly = 4
	OccupancyFull                    = 5
	OccupancyNotAcceptingPassengers  = 6
	OccupancyNoDataAvailable         = 7
	OccupancyNotBoardable            = 8
)

// GTFS-RT Congestion Level values (from protobuf)
const (
	CongestionUnknown          = 0
	CongestionRunningSmoothly  = 1
	CongestionStopAndGo        = 2
	CongestionCongestion       = 3
	CongestionSevereCongestion = 4
)

// ==============================================================================
// HELPER FUNCTIONS (from wrapper.go)
// ==============================================================================

// TripKeyForConverter creates a composite trip key for GTFS static lookups
// Format: {agency}_{startDate}_{tripID} (or just tripID if parts are empty)
func TripKeyForConverter(tripID, agency, startDate string) string {
	key := tripID
	if startDate != "" {
		key = startDate + "_" + key
	}
	if agency != "" {
		key = agency + "_" + key
	}
	return key
}

// translatedStringToText extracts best-effort text from GTFS-RT TranslatedString
// Translation represents a single language translation
type Translation struct {
	Language string
	Text     string
}

// translatedStringToText prefers entries with no language tag, or returns first available
// Currently unused but kept for future multi-language support
func translatedStringToText(translations []Translation) string {
	if len(translations) == 0 {
		return ""
	}
	var first string
	for _, tr := range translations {
		if tr.Language == "" {
			return tr.Text
		}
		if first == "" {
			first = tr.Text
		}
	}
	return first
}

// translatedStringToMap extracts all translations as language -> text map
// Currently unused but kept for future multi-language support
func translatedStringToMap(translations []Translation) map[string]string {
	result := make(map[string]string)
	for _, tr := range translations {
		lang := tr.Language
		if lang == "" {
			lang = "unknown"
		}
		result[lang] = tr.Text
	}
	return result
}

// Suppress unused warnings - these functions are kept for future use
var _ = translatedStringToText
var _ = translatedStringToMap

