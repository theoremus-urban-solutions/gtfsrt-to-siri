package converter

import (
	"math"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

func (c *Converter) buildMVJ(tripID string) siri.MonitoredVehicleJourney {
	agency := c.opts.AgencyID
	startDate := c.gtfsrt.GetStartDateForTrip(tripID)
	tripKey := gtfsrt.TripKeyForConverter(tripID, agency, startDate)

	// Prefer RT route_id; fallback to static lookup by tripID (ALWAYS use plain tripID for static GTFS)
	routeID := c.gtfsrt.GetRouteIDForTrip(tripID)
	if routeID == "" {
		routeID = c.gtfs.GetRouteIDForTrip(tripID)
		if routeID == "" {
			c.warnings.Add(WarningNoRouteID, tripID)
			routeID = "UNKNOWN"
		}
	}
	// LineRef format: {codespace}:Line:{lineid}
	lineRef := routeID
	if agency != "" && routeID != "" {
		lineRef = agency + ":Line:" + routeID
	}
	direction := c.gtfsrt.GetRouteDirectionForTrip(tripID)
	if direction == "" {
		direction = c.gtfs.GetDirectionIDForTrip(tripID)
	}
	// Get origin and destination stops (format: {codespace}:Quay:{stopid})
	originStopID := c.gtfs.GetOriginStopIDForTrip(tripID)
	originStopID = applyFieldMutators(originStopID, c.opts.FieldMutators.OriginRef)
	origin := ""
	if originStopID != "" && agency != "" {
		origin = agency + ":Quay:" + originStopID
	}
	originName := c.gtfs.GetStopName(originStopID)
	if originStopID != "" && originName == "" {
		c.warnings.Add(WarningOriginStopNoName, tripID)
	}

	destStopID := c.gtfs.GetDestinationStopIDForTrip(tripID)
	destStopID = applyFieldMutators(destStopID, c.opts.FieldMutators.DestinationRef)
	dest := ""
	if destStopID != "" && agency != "" {
		dest = agency + ":Quay:" + destStopID
	}
	head := c.gtfs.GetTripHeadsign(tripID)

	// PublishedLineName: try route_short_name, fallback to route_id
	pub := c.gtfs.GetRouteShortName(routeID)
	if pub == "" && routeID != "" && routeID != "UNKNOWN" {
		pub = routeID // Use route_id as fallback
		c.warnings.Add(WarningNoRouteShortName, tripID)
	}

	// VehicleRef format: {codespace}:VehicleRef:{vehicle_id}
	vehRef := ""
	if rawVehicleID := c.gtfsrt.GetVehicleRefForTrip(tripID); rawVehicleID != "" {
		vehRef = agency + ":VehicleRef:" + rawVehicleID
	}
	var bearing *float64
	if b, ok := c.gtfsrt.GetVehicleBearingForTrip(tripID); ok {
		bearing = &b
	}
	var latPtr *float64
	var lonPtr *float64
	if lat, ok := c.gtfsrt.GetVehicleLatForTrip(tripID); ok {
		latPtr = &lat
	}
	if lon, ok := c.gtfsrt.GetVehicleLonForTrip(tripID); ok {
		lonPtr = &lon
	}
	// Snapshot fallbacks when RT missing
	if latPtr == nil || lonPtr == nil {
		if sLat := c.snap.GetLatitude(tripKey); sLat != nil && latPtr == nil {
			latPtr = sLat
		}
		if sLon := c.snap.GetLongitude(tripKey); sLon != nil && lonPtr == nil {
			lonPtr = sLon
		}
	}
	if bearing == nil {
		if sB := c.snap.GetBearing(tripKey); sB != nil {
			bearing = sB
		}
	}
	// Log warning if still no position data
	if latPtr == nil || lonPtr == nil {
		c.warnings.Add(WarningNoLatLon, tripID)
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
		if dep := c.gtfsrt.GetExpectedDepartureTimeAtStopForTrip(tripID, originStopID); dep > 0 {
			originAimed = utils.Iso8601FromUnixSeconds(dep)
		} else if arr := c.gtfsrt.GetExpectedArrivalTimeAtStopForTrip(tripID, originStopID); arr > 0 {
			originAimed = utils.Iso8601FromUnixSeconds(arr)
		} else {
			// Fall back to GTFS static departure time
			if staticDepTime := c.gtfs.GetDepartureTime(tripID, originStopID); staticDepTime != "" {
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

	// Get vehicle speed from GTFS-RT VehiclePosition (in m/s)
	var velocity *int
	if speed, ok := c.gtfsrt.GetVehicleSpeedForTrip(tripID); ok {
		// GTFS-RT speed is in m/s, SIRI Velocity is also in m/s (rounded to int)
		v := int(math.Round(speed))
		velocity = &v
	}

	// Build MonitoredCall (current or next stop) instead of OnwardCalls for VM
	monitoredCall := c.buildMonitoredCall(tripID)

	// Get VehicleMode from route_type (same as ET)
	vehicleMode := ""
	if routeType := c.gtfs.GetRouteType(routeID); routeType >= 0 {
		vehicleMode = mapGTFSRouteTypeToSIRIVehicleMode(routeType)
	}

	// OperatorRef format: {codespace}:Operator:{operator_name}
	operatorRef := agency
	if agencyName := c.gtfs.GetAgencyName(); agencyName != "" {
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
		Velocity:                 velocity, // Speed in m/s from VehiclePosition
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
		return agency + "_" + c.gtfs.GetFullTripIDForTrip(tripKey)
	}
	return c.gtfs.GetFullTripIDForTrip(tripKey)
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
	stops := c.gtfsrt.GetOnwardStopIDsForTrip(tripID)
	if len(stops) == 0 {
		return "PT0S" // No stops, no delay
	}

	currentStopID := stops[0] // First onward stop is current/next

	// Get start date for time conversion
	startDate := c.gtfsrt.GetStartDateForTrip(tripID)
	// If no start_date from GTFS-RT, use today's date
	if startDate == "" {
		startDate = utils.Iso8601DateFromUnixSeconds(c.gtfsrt.GetTimestampForFeedMessage())
		startDate = startDate[0:4] + startDate[5:7] + startDate[8:10] // Convert YYYY-MM-DD to YYYYMMDD
	}

	// Get expected time from GTFS-RT (prefer departure, fallback to arrival)
	var expectedTime int64
	if dep := c.gtfsrt.GetExpectedDepartureTimeAtStopForTrip(tripID, currentStopID); dep > 0 {
		expectedTime = dep
	} else if arr := c.gtfsrt.GetExpectedArrivalTimeAtStopForTrip(tripID, currentStopID); arr > 0 {
		expectedTime = arr
	} else {
		return "PT0S" // No RT data
	}

	// Get scheduled time from GTFS static
	gtfsTime := c.gtfs.GetDepartureTime(tripID, currentStopID)
	if gtfsTime == "" {
		// Try arrival time if departure not available
		gtfsTime = c.gtfs.GetArrivalTime(tripID, currentStopID)
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

// mapOccupancyStatus maps GTFS-RT VehiclePosition occupancy_status to SIRI Occupancy values
func (c *Converter) mapOccupancyStatus(tripID string) string {
	occupancyStatus := c.gtfsrt.GetOccupancyStatusForTrip(tripID)

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
	congestionLevel := c.gtfsrt.GetCongestionLevelForTrip(tripID)

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
	stops := c.gtfsrt.GetOnwardStopIDsForTrip(tripID)
	if len(stops) == 0 {
		c.warnings.Add(WarningNoOnwardStops, tripID)
		return nil
	}

	// Use first onward stop as "current" stop
	currentStopID := stops[0]

	// Check if vehicle is at stop (distance < 50m)
	agency := c.opts.AgencyID
	startDate := c.gtfsrt.GetStartDateForTrip(tripID)
	tripKey := gtfsrt.TripKeyForConverter(tripID, agency, startDate)

	vehKM := c.snap.GetVehicleDistanceAlongRouteInKilometers(tripKey)
	stopKM := c.gtfs.GetStopDistanceAlongRouteForTripInKilometers(tripID, currentStopID)
	distanceToStop := (stopKM - vehKM) * 1000 // meters

	vehicleAtStop := !math.IsNaN(vehKM) && distanceToStop >= -50 && distanceToStop <= 50

	stopName := c.gtfs.GetStopName(currentStopID)
	if stopName == "" {
		c.warnings.Add(WarningMonitoredCallStopNoName, tripID)
	}

	// Get stop order/sequence from GTFS static
	var order *int
	if stopSeq := c.gtfs.TripStopSeq[tripID]; len(stopSeq) > 0 {
		for i, stopID := range stopSeq {
			if stopID == currentStopID {
				orderVal := i + 1 // 1-based index
				order = &orderVal
				break
			}
		}
	}

	// Format StopPointRef as {codespace}:Quay:{stopid}
	currentStopID = applyFieldMutators(currentStopID, c.opts.FieldMutators.StopPointRef)
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
