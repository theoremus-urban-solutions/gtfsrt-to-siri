package converter

import (
	"fmt"
	"log"
	"strings"
)

// Warning type constants
const (
	// VM warnings
	WarningNoRouteID               = "no_route_id"
	WarningOriginStopNoName        = "origin_stop_no_name"
	WarningNoRouteShortName        = "no_route_short_name"
	WarningNoLatLon                = "no_lat_lon"
	WarningNoOnwardStops           = "no_onward_stops"
	WarningMonitoredCallStopNoName = "monitored_call_stop_no_name"

	// ET warnings
	WarningNoStartDate       = "no_start_date"
	WarningTripNotInStatic   = "trip_not_in_static"
	WarningNoRouteType       = "no_route_type"
	WarningDestStopNoName    = "dest_stop_no_name"
	WarningNoStaticTimes     = "no_static_times"
	WarningStopNoName        = "stop_no_name"
	WarningNoArrivalTime     = "no_arrival_time"
	WarningNoDepartureTime   = "no_departure_time"
	WarningNoStopTimeUpdates = "no_stop_time_updates"

	// SX warnings
	WarningNoSummary     = "no_summary"
	WarningNoDescription = "no_description"
	WarningStopNotFound  = "stop_not_found"
)

// warningInfo holds aggregated information about a specific warning type
type warningInfo struct {
	count    int
	examples []string
}

// WarningAggregator collects warnings during conversion and outputs consolidated summaries
type WarningAggregator struct {
	warnings map[string]*warningInfo
}

// NewWarningAggregator creates a new warning aggregator
func NewWarningAggregator() *WarningAggregator {
	return &WarningAggregator{
		warnings: make(map[string]*warningInfo),
	}
}

// Add records a warning occurrence with an example ID
func (w *WarningAggregator) Add(warningType, exampleID string) {
	if w.warnings[warningType] == nil {
		w.warnings[warningType] = &warningInfo{
			examples: make([]string, 0, 3),
		}
	}

	info := w.warnings[warningType]
	info.count++

	// Store up to 3 examples
	if len(info.examples) < 3 {
		info.examples = append(info.examples, exampleID)
	}
}

// LogAll outputs all collected warnings in consolidated format
func (w *WarningAggregator) LogAll(feedModule, agencyID string) {
	if len(w.warnings) == 0 {
		return
	}

	for warningType, info := range w.warnings {
		message := w.formatWarningMessage(warningType, feedModule, agencyID, info)
		log.Printf("%s", message)
	}
}

// formatWarningMessage creates a human-readable warning message
func (w *WarningAggregator) formatWarningMessage(warningType, feedModule, agencyID string, info *warningInfo) string {
	var description, action string

	switch warningType {
	case WarningNoRouteID:
		description = "missing route_id fields"
		action = "Building SIRI output with fallback 'UNKNOWN'"
	case WarningOriginStopNoName:
		description = "origin stops with no name in static GTFS"
		action = "Building SIRI output with empty origin name"
	case WarningNoRouteShortName:
		description = "routes with no route_short_name"
		action = "Using route_id as fallback for PublishedLineName"
	case WarningNoLatLon:
		description = "trips with no lat/lon in GTFS-RT or snapshot"
		action = "Building SIRI output without VehicleLocation"
	case WarningNoOnwardStops:
		description = "trips with no onward stops in GTFS-RT"
		action = "Building SIRI output without MonitoredCall"
	case WarningMonitoredCallStopNoName:
		description = "monitored call stops with no name in static GTFS"
		action = "Building SIRI output with empty stop name"
	case WarningNoStartDate:
		description = "trips with no start_date"
		action = "Using current date as fallback"
	case WarningTripNotInStatic:
		description = "trips not found in static GTFS"
		action = "Building minimal calls from GTFS-RT only"
	case WarningNoRouteType:
		description = "routes with no route_type in static GTFS"
		action = "Building SIRI output without VehicleMode"
	case WarningDestStopNoName:
		description = "destination stops with no name in static GTFS"
		action = "Building SIRI output with empty destination name"
	case WarningNoStaticTimes:
		description = "stops with no static times in GTFS"
		action = "Using only GTFS-RT times where available"
	case WarningStopNoName:
		description = "stops with no name in static GTFS"
		action = "Building SIRI output with empty stop name"
	case WarningNoArrivalTime:
		description = "stops with no arrival time (neither RT nor static)"
		action = "Building SIRI output without arrival time"
	case WarningNoDepartureTime:
		description = "stops with no departure time (neither RT nor static)"
		action = "Building SIRI output without departure time"
	case WarningNoStopTimeUpdates:
		description = "trips with no stop_time_updates in GTFS-RT"
		action = "Building SIRI output without calls"
	case WarningNoSummary:
		description = "alerts with no header_text/summary"
		action = "Building SIRI output with empty summary"
	case WarningNoDescription:
		description = "alerts with no description_text"
		action = "Building SIRI output with empty description"
	case WarningStopNotFound:
		description = "stops not found in static GTFS"
		action = "Building SIRI output with stop reference only"
	default:
		description = "unknown issue"
		action = "Building SIRI output with fallback behavior"
	}

	examplesStr := strings.Join(info.examples, ", ")

	return fmt.Sprintf("Feed %s for agency %s has %s (%d occurrences). %s. Examples: %s",
		feedModule, agencyID, description, info.count, action, examplesStr)
}
