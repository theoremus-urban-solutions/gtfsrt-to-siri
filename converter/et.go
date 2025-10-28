package converter

import (
	"fmt"
	"time"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

// BuildEstimatedTimetable converts GTFS-RT data to SIRI ET format
func (c *Converter) BuildEstimatedTimetable() siri.EstimatedTimetableDelivery {
	timestamp := c.gtfsrt.GetTimestampForFeedMessage()
	now := timestamp
	agencyID := c.opts.AgencyID
	if agencyID == "" {
		agencyID = "UNKNOWN"
	}

	// Get trips from TripUpdates only (ET should only include trips with trip update data)
	allTrips := c.gtfsrt.GetTripsFromTripUpdates()
	journeys := make([]siri.EstimatedVehicleJourney, 0, len(allTrips))

	for _, tripID := range allTrips {
		journey := c.buildEstimatedVehicleJourney(tripID, now, agencyID)
		if journey != nil {
			journeys = append(journeys, *journey)
		}
	}

	frame := siri.EstimatedJourneyVersionFrame{
		RecordedAtTime:          utils.Iso8601ExtendedFromUnixSeconds(timestamp),
		EstimatedVehicleJourney: journeys,
	}

	// Log consolidated warnings
	c.warnings.LogAll("TU->ET", agencyID)

	return siri.EstimatedTimetableDelivery{
		Version:                      "2.0",
		ResponseTimestamp:            utils.Iso8601ExtendedFromUnixSeconds(timestamp),
		EstimatedJourneyVersionFrame: []siri.EstimatedJourneyVersionFrame{frame},
	}
}

func (c *Converter) buildEstimatedVehicleJourney(tripID string, now int64, agencyID string) *siri.EstimatedVehicleJourney {
	// Get route and direction - try GTFS-RT first, then fall back to static GTFS
	// IMPORTANT: Always use plain tripID for GTFS static lookups (never composite keys)
	routeID := c.gtfsrt.GetRouteIDForTrip(tripID)
	if routeID == "" {
		// Try to get from static GTFS trips.txt using plain trip_id
		routeID = c.gtfs.GetRouteIDForTrip(tripID)
		if routeID == "" {
			c.warnings.Add(WarningNoRouteID, tripID)
			routeID = "UNKNOWN"
		}
	}

	directionID := c.gtfsrt.GetRouteDirectionForTrip(tripID)
	if directionID == "" {
		directionID = "0"
	}

	// Get start date for SIRI output
	startDate := c.gtfsrt.GetStartDateForTrip(tripID)

	// Build siri.FramedVehicleJourneyRef
	dataFrameRef := startDate
	if dataFrameRef == "" {
		dataFrameRef = utils.Iso8601DateFromUnixSeconds(now)
		c.warnings.Add(WarningNoStartDate, tripID)
	}
	datedVehicleJourneyRef := agencyID + ":ServiceJourney:" + tripID

	// Get vehicle ref if available and format as {codespace}:VehicleRef:{vehicle_id}
	vehicleRef := ""
	if rawVehicleID := c.gtfsrt.GetVehicleRefForTrip(tripID); rawVehicleID != "" {
		vehicleRef = agencyID + ":VehicleRef:" + rawVehicleID
	}

	// Get complete stop sequence from GTFS static - ALWAYS use plain tripID for static GTFS
	stopSequence := c.gtfs.TripStopSeq[tripID]

	var recordedCalls []siri.RecordedCall
	var estimatedCalls []siri.EstimatedCall

	if len(stopSequence) == 0 {
		// Trip exists in GTFS-RT but not in GTFS static - build calls from RT only
		c.warnings.Add(WarningTripNotInStatic, tripID)
		recordedCalls, estimatedCalls = c.buildCallSequenceFromRTOnly(tripID, now)
	} else {
		// Split into siri.RecordedCalls and siri.EstimatedCalls (always use plain tripID for static GTFS)
		recordedCalls, estimatedCalls = c.buildCallSequence(tripID, tripID, stopSequence, now)
	}

	// Get VehicleMode from route_type
	vehicleMode := ""
	if routeType := c.gtfs.GetRouteType(routeID); routeType > 0 {
		vehicleMode = mapGTFSRouteTypeToSIRIVehicleMode(routeType)
	} else if routeID != "UNKNOWN" {
		c.warnings.Add(WarningNoRouteType, tripID+":"+routeID)
	}

	// Get Origin and Destination names from first/last stop in calls
	originName := ""
	destinationName := ""
	if len(stopSequence) > 0 {
		originName = c.gtfs.GetStopName(stopSequence[0])
		destinationName = c.gtfs.GetStopName(stopSequence[len(stopSequence)-1])
		if originName == "" {
			c.warnings.Add(WarningOriginStopNoName, tripID)
		}
		if destinationName == "" {
			c.warnings.Add(WarningDestStopNoName, tripID)
		}
	}

	// OperatorRef from agency_name
	operatorRef := agencyID
	if agencyName := c.gtfs.GetAgencyName(); agencyName != "" {
		operatorRef = agencyID + ":Operator:" + agencyName
	}

	// Monitored: true if trip is currently ongoing (has both past and future stops)
	monitored := len(recordedCalls) > 0 && len(estimatedCalls) > 0

	journey := &siri.EstimatedVehicleJourney{
		RecordedAtTime: utils.Iso8601ExtendedFromUnixSeconds(now),
		LineRef:        agencyID + ":Line:" + routeID,
		VehicleRef:     vehicleRef,
		DirectionRef:   directionID,
		FramedVehicleJourneyRef: siri.FramedVehicleJourneyRef{
			DataFrameRef:           dataFrameRef,
			DatedVehicleJourneyRef: datedVehicleJourneyRef,
		},
		VehicleMode:            vehicleMode,
		OriginName:             originName,
		DestinationName:        destinationName,
		Monitored:              monitored,
		DataSource:             agencyID,
		OperatorRef:            operatorRef,
		RecordedCalls:          recordedCalls,
		EstimatedCalls:         estimatedCalls,
		IsCompleteStopSequence: true,
	}

	return journey
}

func (c *Converter) buildCallSequence(tripID, gtfsLookupKey string, stopSequence []string, now int64) ([]siri.RecordedCall, []siri.EstimatedCall) {
	recordedCalls := []siri.RecordedCall{}
	estimatedCalls := []siri.EstimatedCall{}
	agencyID := c.opts.AgencyID
	if agencyID == "" {
		agencyID = "UNKNOWN"
	}

	// Get start_date for time conversion
	startDate := c.gtfsrt.GetStartDateForTrip(tripID)
	// If no start_date from GTFS-RT, use today's date
	if startDate == "" {
		startDate = time.Unix(now, 0).Format("20060102")
	}

	for order, stopID := range stopSequence {
		// Get real-time arrival/departure times
		rtArrival := c.gtfsrt.GetExpectedArrivalTimeAtStopForTrip(tripID, stopID)
		rtDeparture := c.gtfsrt.GetExpectedDepartureTimeAtStopForTrip(tripID, stopID)

		// Get static GTFS times using gtfsLookupKey (the key that successfully found stopSequence)
		staticArrivalStr := c.gtfs.GetArrivalTime(gtfsLookupKey, stopID)
		staticDepartureStr := c.gtfs.GetDepartureTime(gtfsLookupKey, stopID)
		staticArrival := gtfsTimeToUnixTimestamp(staticArrivalStr, startDate)
		staticDeparture := gtfsTimeToUnixTimestamp(staticDepartureStr, startDate)

		// Log warnings for missing static times
		if staticArrivalStr == "" && staticDepartureStr == "" {
			c.warnings.Add(WarningNoStaticTimes, tripID+":"+stopID)
		}

		// Determine if this is a past or future stop
		isPastStop := false
		if rtDeparture > 0 && rtDeparture < now {
			isPastStop = true
		} else if rtArrival > 0 && rtArrival < now-60 { // Allow 60s grace period
			isPastStop = true
		}

		// Get stop name
		stopName := c.gtfs.GetStopName(stopID)
		if stopName == "" {
			c.warnings.Add(WarningStopNoName, tripID+":"+stopID)
		}

		// Format StopPointRef as {codespace}:Quay:{stop_id}, then apply field mutators
		stopPointRef := applyFieldMutators(agencyID+":Quay:"+stopID, c.opts.FieldMutators.StopPointRef)

		// Check if cancelled (schedule_relationship = 1 SKIPPED)
		schedRel := c.gtfsrt.GetScheduleRelationshipForStop(tripID, stopID)
		isCancelled := schedRel == 1

		// Check if request stop (pickup_type or drop_off_type = 2 or 3) using gtfsLookupKey
		pickupType := c.gtfs.GetPickupType(gtfsLookupKey, stopID)
		dropOffType := c.gtfs.GetDropOffType(gtfsLookupKey, stopID)
		isRequestStop := pickupType == 2 || pickupType == 3 || dropOffType == 2 || dropOffType == 3

		if isPastStop {
			// siri.RecordedCall
			call := siri.RecordedCall{
				StopPointRef:  stopPointRef,
				Order:         order + 1,
				StopPointName: stopName,
				Cancellation:  isCancelled,
				RequestStop:   isRequestStop,
			}

			// Set aimed times from static GTFS
			if staticArrival > 0 {
				call.AimedArrivalTime = utils.Iso8601ExtendedFromUnixSeconds(staticArrival)
			}
			if staticDeparture > 0 {
				call.AimedDepartureTime = utils.Iso8601ExtendedFromUnixSeconds(staticDeparture)
			}

			// Set actual times from GTFS-RT
			if rtArrival > 0 {
				call.ActualArrivalTime = utils.Iso8601ExtendedFromUnixSeconds(rtArrival)
			} else if staticArrival == 0 {
				c.warnings.Add(WarningNoArrivalTime, tripID+":"+stopID)
			}
			if rtDeparture > 0 {
				call.ActualDepartureTime = utils.Iso8601ExtendedFromUnixSeconds(rtDeparture)
			} else if staticDeparture == 0 {
				c.warnings.Add(WarningNoDepartureTime, tripID+":"+stopID)
			}

			recordedCalls = append(recordedCalls, call)
		} else {
			// siri.EstimatedCall
			call := siri.EstimatedCall{
				StopPointRef:  stopPointRef,
				Order:         order + 1,
				StopPointName: stopName,
				Cancellation:  isCancelled,
				RequestStop:   isRequestStop,
			}

			// Set aimed times from static GTFS
			if staticArrival > 0 {
				call.AimedArrivalTime = utils.Iso8601ExtendedFromUnixSeconds(staticArrival)
			}
			if staticDeparture > 0 {
				call.AimedDepartureTime = utils.Iso8601ExtendedFromUnixSeconds(staticDeparture)
			}

			// Set expected times and status - use RT if available, otherwise fall back to static
			if staticArrival > 0 {
				if rtArrival > 0 {
					call.ExpectedArrivalTime = utils.Iso8601ExtendedFromUnixSeconds(rtArrival)
					call.ArrivalStatus = calculateStatus(rtArrival, staticArrival)
				} else {
					// No real-time data, use static time
					call.ExpectedArrivalTime = utils.Iso8601ExtendedFromUnixSeconds(staticArrival)
					call.ArrivalStatus = "onTime"
				}
			} else if rtArrival > 0 {
				// No static time, but we have RT time - use it
				call.ExpectedArrivalTime = utils.Iso8601ExtendedFromUnixSeconds(rtArrival)
				call.ArrivalStatus = "onTime"
			} else {
				c.warnings.Add(WarningNoArrivalTime, tripID+":"+stopID)
			}

			if staticDeparture > 0 {
				if rtDeparture > 0 {
					call.ExpectedDepartureTime = utils.Iso8601ExtendedFromUnixSeconds(rtDeparture)
					call.DepartureStatus = calculateStatus(rtDeparture, staticDeparture)
				} else {
					// No real-time data, use static time
					call.ExpectedDepartureTime = utils.Iso8601ExtendedFromUnixSeconds(staticDeparture)
					call.DepartureStatus = "onTime"
				}
			} else if rtDeparture > 0 {
				// No static time, but we have RT time - use it
				call.ExpectedDepartureTime = utils.Iso8601ExtendedFromUnixSeconds(rtDeparture)
				call.DepartureStatus = "onTime"
			} else {
				c.warnings.Add(WarningNoDepartureTime, tripID+":"+stopID)
			}

			estimatedCalls = append(estimatedCalls, call)
		}
	}

	return recordedCalls, estimatedCalls
}

// buildCallSequenceFromRTOnly builds minimal call sequence using only GTFS-RT data
// when static GTFS data is unavailable. This allows conversion to continue with
// whatever real-time data we have.
func (c *Converter) buildCallSequenceFromRTOnly(tripID string, now int64) ([]siri.RecordedCall, []siri.EstimatedCall) {
	recordedCalls := []siri.RecordedCall{}
	estimatedCalls := []siri.EstimatedCall{}
	agencyID := c.opts.AgencyID
	if agencyID == "" {
		agencyID = "UNKNOWN"
	}

	// Get stop sequence from GTFS-RT stop_time_updates
	rtStopSequence := c.gtfsrt.GetOnwardStopIDsForTrip(tripID)
	if len(rtStopSequence) == 0 {
		c.warnings.Add(WarningNoStopTimeUpdates, tripID)
		return recordedCalls, estimatedCalls
	}

	for order, stopID := range rtStopSequence {
		// Get real-time arrival/departure times
		rtArrival := c.gtfsrt.GetExpectedArrivalTimeAtStopForTrip(tripID, stopID)
		rtDeparture := c.gtfsrt.GetExpectedDepartureTimeAtStopForTrip(tripID, stopID)

		// Determine if this is a past or future stop
		isPastStop := false
		if rtDeparture > 0 && rtDeparture < now {
			isPastStop = true
		} else if rtArrival > 0 && rtArrival < now-60 { // Allow 60s grace period
			isPastStop = true
		}

		// Format StopPointRef as {codespace}:Quay:{stop_id}, then apply field mutators
		stopPointRef := applyFieldMutators(agencyID+":Quay:"+stopID, c.opts.FieldMutators.StopPointRef)

		// Check if cancelled (schedule_relationship = 1 SKIPPED)
		schedRel := c.gtfsrt.GetScheduleRelationshipForStop(tripID, stopID)
		isCancelled := schedRel == 1

		if isPastStop {
			// siri.RecordedCall
			call := siri.RecordedCall{
				StopPointRef:  stopPointRef,
				Order:         order + 1,
				StopPointName: "", // No static data available
				Cancellation:  isCancelled,
				RequestStop:   false, // No static data available
			}

			// Set actual times from GTFS-RT (no aimed times without static data)
			if rtArrival > 0 {
				call.ActualArrivalTime = utils.Iso8601ExtendedFromUnixSeconds(rtArrival)
			}
			if rtDeparture > 0 {
				call.ActualDepartureTime = utils.Iso8601ExtendedFromUnixSeconds(rtDeparture)
			}

			recordedCalls = append(recordedCalls, call)
		} else {
			// siri.EstimatedCall
			call := siri.EstimatedCall{
				StopPointRef:  stopPointRef,
				Order:         order + 1,
				StopPointName: "", // No static data available
				Cancellation:  isCancelled,
				RequestStop:   false, // No static data available
			}

			// Set expected times from GTFS-RT (no aimed times without static data)
			if rtArrival > 0 {
				call.ExpectedArrivalTime = utils.Iso8601ExtendedFromUnixSeconds(rtArrival)
				// Without static schedule, we can't determine status
				call.ArrivalStatus = "onTime"
			}
			if rtDeparture > 0 {
				call.ExpectedDepartureTime = utils.Iso8601ExtendedFromUnixSeconds(rtDeparture)
				// Without static schedule, we can't determine status
				call.DepartureStatus = "onTime"
			}

			estimatedCalls = append(estimatedCalls, call)
		}
	}

	return recordedCalls, estimatedCalls
}

// calculateStatus determines the status based on delay
func calculateStatus(expectedTime, aimedTime int64) string {
	delay := expectedTime - aimedTime

	if delay <= -60 {
		return "early"
	} else if delay < 60 {
		return "onTime"
	} else {
		return "delayed"
	}
}

// gtfsTimeToUnixTimestamp converts a GTFS time string (HH:MM:SS) and start_date (YYYYMMDD) to Unix timestamp
// GTFS times can be > 24h (e.g., 25:30:00 for 1:30 AM next day)
func gtfsTimeToUnixTimestamp(gtfsTime, startDate string) int64 {
	if gtfsTime == "" || startDate == "" {
		return 0
	}
	// Parse start_date: YYYYMMDD
	if len(startDate) != 8 {
		return 0
	}
	year := startDate[0:4]
	month := startDate[4:6]
	day := startDate[6:8]

	// Parse time: HH:MM:SS
	var hour, min, sec int
	if _, err := fmt.Sscanf(gtfsTime, "%d:%d:%d", &hour, &min, &sec); err != nil {
		return 0
	}

	// Build date string and parse in local timezone (not UTC)
	// GTFS static times are in local time, not UTC
	dateStr := fmt.Sprintf("%s-%s-%sT00:00:00", year, month, day)
	t, err := time.ParseInLocation("2006-01-02T15:04:05", dateStr, time.Local)
	if err != nil {
		return 0
	}

	// Add hours, minutes, seconds (handles > 24h times)
	t = t.Add(time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec)*time.Second)
	return t.Unix()
}

// mapGTFSRouteTypeToSIRIVehicleMode maps GTFS route_type to SIRI VehicleMode
// See: https://gtfs.org/schedule/reference/#routestxt
func mapGTFSRouteTypeToSIRIVehicleMode(routeType int) string {
	switch routeType {
	case 0:
		return "tram"
	case 1:
		return "metro"
	case 2:
		return "rail"
	case 3:
		return "bus"
	case 4:
		return "ferry"
	case 5:
		return "cableTram"
	case 6:
		return "aerialLift"
	case 7:
		return "funicular"
	case 11:
		return "trolleybus"
	case 12:
		return "monorail"
	default:
		return "bus" // Default fallback
	}
}
