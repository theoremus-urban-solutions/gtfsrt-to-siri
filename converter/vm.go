package converter

import (
	"math"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

func (c *Converter) buildMVJ(tripID string) siri.MonitoredVehicleJourney {
	agency := c.Cfg.GTFS.AgencyID
	startDate := c.GTFSRT.GetStartDateForTrip(tripID)
	tripKey := gtfsrt.TripKeyForConverter(tripID, agency, startDate)

	// Prefer RT route_id; fallback to static lookup by tripKey
	routeID := c.GTFSRT.GetRouteIDForTrip(tripID)
	if routeID == "" {
		routeID = c.GTFS.GetRouteIDForTrip(tripKey)
	}
	// LineRef format: {codespace}:Line:{lineid}
	lineRef := routeID
	if agency != "" && routeID != "" {
		lineRef = agency + ":Line:" + routeID
	}
	direction := c.GTFSRT.GetRouteDirectionForTrip(tripID)
	if direction == "" {
		direction = c.GTFS.GetDirectionIDForTrip(tripKey)
	}
	// Get origin and destination stops (format: {codespace}:Quay:{stopid})
	originStopID := c.GTFS.GetOriginStopIDForTrip(tripKey)
	originStopID = applyFieldMutators(originStopID, c.Cfg.Converter.FieldMutators.OriginRef)
	origin := ""
	if originStopID != "" && agency != "" {
		origin = agency + ":Quay:" + originStopID
	}
	originName := c.GTFS.GetStopName(originStopID)

	destStopID := c.GTFS.GetDestinationStopIDForTrip(tripKey)
	destStopID = applyFieldMutators(destStopID, c.Cfg.Converter.FieldMutators.DestinationRef)
	dest := ""
	if destStopID != "" && agency != "" {
		dest = agency + ":Quay:" + destStopID
	}
	head := c.GTFS.GetTripHeadsign(tripKey)
	pub := c.GTFS.GetRouteShortName(routeID)

	// VehicleRef format: {codespace}:VehicleRef:{vehicle_id}
	vehRef := ""
	if rawVehicleID := c.GTFSRT.GetVehicleRefForTrip(tripID); rawVehicleID != "" {
		vehRef = agency + ":VehicleRef:" + rawVehicleID
	}
	var bearing *float64
	if b, ok := c.GTFSRT.GetVehicleBearingForTrip(tripID); ok {
		bearing = &b
	}
	var latPtr *float64
	var lonPtr *float64
	if lat, ok := c.GTFSRT.GetVehicleLatForTrip(tripID); ok {
		latPtr = &lat
	}
	if lon, ok := c.GTFSRT.GetVehicleLonForTrip(tripID); ok {
		lonPtr = &lon
	}
	// Snapshot fallbacks when RT missing
	if latPtr == nil || lonPtr == nil {
		if sLat := c.Snap.GetLatitude(tripKey); sLat != nil && latPtr == nil {
			latPtr = sLat
		}
		if sLon := c.Snap.GetLongitude(tripKey); sLon != nil && lonPtr == nil {
			lonPtr = sLon
		}
	}
	if bearing == nil {
		if sB := c.Snap.GetBearing(tripKey); sB != nil {
			bearing = sB
		}
	}

	// siri.FramedVehicleJourneyRef with DataFrameRef (YYYY-MM-DD) and DatedVehicleJourneyRef ({codespace}:ServiceJourney:{tripID})
	dataFrameRef := startDate
	if len(startDate) == 8 { // YYYYMMDD -> YYYY-MM-DD
		dataFrameRef = startDate[:4] + "-" + startDate[4:6] + "-" + startDate[6:8]
	}
	datedVehicleJourneyRef := agency + ":ServiceJourney:" + tripID

	// OriginAimedDepartureTime fallback order: RT dep at origin, else RT arr at origin, else GTFS static departure time
	originAimed := ""
	if originStopID != "" {
		if dep := c.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(tripID, originStopID); dep > 0 {
			originAimed = utils.Iso8601FromUnixSeconds(dep)
		} else if arr := c.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(tripID, originStopID); arr > 0 {
			originAimed = utils.Iso8601FromUnixSeconds(arr)
		} else {
			// Fall back to GTFS static departure time
			if staticDepTime := c.GTFS.GetDepartureTime(tripID, originStopID); staticDepTime != "" {
				// Convert HH:MM:SS to ISO8601 timestamp using the start date
				originAimed = utils.Iso8601FromGTFSTimeAndDate(staticDepTime, startDate)
			}
		}
	}

	// Calculate delay (SIRI-VM spec: required)
	delay := c.calculateDelay(tripID)

	// Map occupancy status from GTFS-RT TripUpdate
	occupancy := c.mapOccupancyStatus(tripID)

	// Map congestion level from GTFS-RT VehiclePosition
	inCongestion := c.mapCongestionLevel(tripID)

	// Build MonitoredCall (current or next stop) instead of OnwardCalls for VM
	monitoredCall := c.buildMonitoredCall(tripID)

	// Get VehicleMode from route_type (same as ET)
	vehicleMode := ""
	if routeType := c.GTFS.GetRouteType(routeID); routeType >= 0 {
		vehicleMode = mapGTFSRouteTypeToSIRIVehicleMode(routeType)
	}

	// OperatorRef format: {codespace}:Operator:{operator_name}
	operatorRef := agency
	if agencyName := c.GTFS.GetAgencyName(); agencyName != "" {
		operatorRef = agency + ":Operator:" + agencyName
	}

	return siri.MonitoredVehicleJourney{
		LineRef:                  lineRef,
		DirectionRef:             direction,
		FramedVehicleJourneyRef:  siri.FramedVehicleJourneyRef{DataFrameRef: dataFrameRef, DatedVehicleJourneyRef: datedVehicleJourneyRef},
		VehicleMode:              vehicleMode,
		PublishedLineName:        pub,
		OperatorRef:              operatorRef,
		OriginRef:                origin,
		OriginName:               originName,
		DestinationRef:           dest,
		DestinationName:          head,
		OriginAimedDepartureTime: originAimed,
		SituationRef:             nil,
		Monitored:                true,
		InCongestion:             inCongestion,
		DataSource:               agency, // SIRI-VM spec: required codespace
		VehicleLocation:          siri.VehicleLocation{Latitude: latPtr, Longitude: lonPtr},
		Bearing:                  bearing,
		Occupancy:                occupancy,
		Delay:                    delay,  // SIRI-VM spec: required
		VehicleRef:               vehRef, // SIRI-VM spec: {codespace}:VehicleRef:{vehicle_id}
		ProgressRate:             nil,
		ProgressStatus:           nil,
		MonitoredCall:            monitoredCall, // SIRI-VM spec: current/previous stop
		IsCompleteStopSequence:   false,         // SIRI-VM spec: required, always false
		OnwardCalls:              nil,           // Remove OnwardCalls for VM (belongs in ET)
	}
}

// CfGDatedVehicleJourneyRef returns a DatedVehicleJourneyRef; concat agency + full trip id based on strategy
func (c *Converter) CfGDatedVehicleJourneyRef(tripKey, agency string) string {
	if agency != "" {
		return agency + "_" + c.GTFS.GetFullTripIDForTrip(tripKey)
	}
	return c.GTFS.GetFullTripIDForTrip(tripKey)
}

// applyFieldMutators applies [from,to] pairs to a reference value
func applyFieldMutators(value string, mapping []string) string {
	if len(mapping) < 2 {
		return value
	}
	for i := 0; i+1 < len(mapping); i += 2 {
		from := mapping[i]
		to := mapping[i+1]
		if value == from {
			return to
		}
	}
	return value
}

// calculateDelay calculates delay as ISO 8601 duration string (SIRI-VM spec: required)
// Compares GTFS-RT expected time with GTFS static scheduled time for the next/current stop
func (c *Converter) calculateDelay(tripID string) string {
	// Get the next/current stop from GTFS-RT
	stops := c.GTFSRT.GetOnwardStopIDsForTrip(tripID)
	if len(stops) == 0 {
		return "PT0S" // No stops, no delay
	}

	currentStopID := stops[0] // First onward stop is current/next

	// Get start date for time conversion
	startDate := c.GTFSRT.GetStartDateForTrip(tripID)
	// If no start_date from GTFS-RT, use today's date
	if startDate == "" {
		startDate = utils.Iso8601DateFromUnixSeconds(c.GTFSRT.GetTimestampForFeedMessage())
		startDate = startDate[0:4] + startDate[5:7] + startDate[8:10] // Convert YYYY-MM-DD to YYYYMMDD
	}

	// Get expected time from GTFS-RT (prefer departure, fallback to arrival)
	var expectedTime int64
	if dep := c.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(tripID, currentStopID); dep > 0 {
		expectedTime = dep
	} else if arr := c.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(tripID, currentStopID); arr > 0 {
		expectedTime = arr
	} else {
		return "PT0S" // No RT data
	}

	// Get scheduled time from GTFS static
	gtfsTime := c.GTFS.GetDepartureTime(tripID, currentStopID)
	if gtfsTime == "" {
		// Try arrival time if departure not available
		gtfsTime = c.GTFS.GetArrivalTime(tripID, currentStopID)
	}
	if gtfsTime == "" || startDate == "" {
		return "PT0S" // No static data
	}

	// Convert GTFS static time to Unix seconds
	scheduledTime := utils.ParseGTFSTimeToUnixSeconds(gtfsTime, startDate)
	if scheduledTime == 0 {
		return "PT0S" // Parsing failed
	}

	// Calculate delay (positive = late, negative = early)
	delaySeconds := expectedTime - scheduledTime

	return utils.FormatDelayAsISO8601Duration(delaySeconds)
}

// mapOccupancyStatus maps GTFS-RT TripUpdate occupancy_status to SIRI Occupancy values
func (c *Converter) mapOccupancyStatus(tripID string) string {
	occupancyStatus := c.GTFSRT.GetOccupancyStatusForTrip(tripID)

	switch occupancyStatus {
	case 0, 1: // EMPTY, MANY_SEATS_AVAILABLE
		return "manySeatsAvailable"
	case 2: // FEW_SEATS_AVAILABLE
		return "seatsAvailable"
	case 3: // STANDING_ROOM_ONLY
		return "standingAvailable"
	case 4: // CRUSHED_STANDING_ROOM_ONLY
		return "standingAvailable"
	case 5: // FULL
		return "full"
	case 6, 8: // NOT_ACCEPTING_PASSENGERS, NOT_BOARDABLE
		return "notAcceptingPassengers"
	case 7: // NO_DATA_AVAILABLE
		return "unknown"
	default:
		return "" // Empty string if no data
	}
}

// mapCongestionLevel maps GTFS-RT VehiclePosition congestion_level to SIRI InCongestion boolean
func (c *Converter) mapCongestionLevel(tripID string) *bool {
	congestionLevel := c.GTFSRT.GetCongestionLevelForTrip(tripID)

	// If no congestion data available, return nil (omit field)
	if congestionLevel < 0 {
		return nil
	}

	// UNKNOWN_CONGESTION_LEVEL(0) or RUNNING_SMOOTHLY(1) => false
	// STOP_AND_GO(2), CONGESTION(3), SEVERE_CONGESTION(4) => true
	inCongestion := congestionLevel >= 2
	return &inCongestion
}

// buildMonitoredCall builds MonitoredCall for current/next stop (SIRI-VM spec)
func (c *Converter) buildMonitoredCall(tripID string) *siri.MonitoredCall {
	stops := c.GTFSRT.GetOnwardStopIDsForTrip(tripID)
	if len(stops) == 0 {
		return nil
	}

	// Use first onward stop as "current" stop
	currentStopID := stops[0]

	// Check if vehicle is at stop (distance < 50m)
	agency := c.Cfg.GTFS.AgencyID
	startDate := c.GTFSRT.GetStartDateForTrip(tripID)
	tripKey := gtfsrt.TripKeyForConverter(tripID, agency, startDate)

	vehKM := c.Snap.GetVehicleDistanceAlongRouteInKilometers(tripKey)
	stopKM := c.GTFS.GetStopDistanceAlongRouteForTripInKilometers(tripKey, currentStopID)
	distanceToStop := (stopKM - vehKM) * 1000 // meters

	vehicleAtStop := !math.IsNaN(vehKM) && distanceToStop >= -50 && distanceToStop <= 50

	stopName := c.GTFS.GetStopName(currentStopID)

	// Get stop order/sequence from GTFS static
	var order *int
	if stopSeq := c.GTFS.TripStopSeq[tripKey]; len(stopSeq) > 0 {
		for i, stopID := range stopSeq {
			if stopID == currentStopID {
				orderVal := i + 1 // 1-based index
				order = &orderVal
				break
			}
		}
	}

	// Format StopPointRef as {codespace}:Quay:{stopid}
	currentStopID = applyFieldMutators(currentStopID, c.Cfg.Converter.FieldMutators.StopPointRef)
	stopPointRef := ""
	if currentStopID != "" && agency != "" {
		stopPointRef = agency + ":Quay:" + currentStopID
	}

	return &siri.MonitoredCall{
		StopPointRef:  stopPointRef,
		Order:         order,
		StopPointName: stopName,
		VehicleAtStop: &vehicleAtStop,
	}
}
