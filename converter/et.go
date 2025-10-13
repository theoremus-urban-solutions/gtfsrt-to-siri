package converter

import (
	"fmt"
	"time"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
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

	return siri.EstimatedTimetableDelivery{
		Version:                      "2.0",
		ResponseTimestamp:            utils.Iso8601ExtendedFromUnixSeconds(timestamp),
		EstimatedJourneyVersionFrame: []siri.EstimatedJourneyVersionFrame{frame},
	}
}

func (c *Converter) buildEstimatedVehicleJourney(tripID string, now int64, agencyID string) *siri.EstimatedVehicleJourney {
	// Get route and direction
	routeID := c.gtfsrt.GetRouteIDForTrip(tripID)
	if routeID == "" {
		return nil
	}

	directionID := c.gtfsrt.GetRouteDirectionForTrip(tripID)
	if directionID == "" {
		directionID = "0"
	}

	// Get trip key
	startDate := c.gtfsrt.GetStartDateForTrip(tripID)
	tripKey := gtfsrt.TripKeyForConverter(tripID, agencyID, startDate)

	// Build siri.FramedVehicleJourneyRef
	dataFrameRef := startDate
	if dataFrameRef == "" {
		dataFrameRef = utils.Iso8601DateFromUnixSeconds(now)
	}
	datedVehicleJourneyRef := agencyID + ":ServiceJourney:" + tripID

	// Get vehicle ref if available and format as {codespace}:VehicleRef:{vehicle_id}
	vehicleRef := ""
	if rawVehicleID := c.gtfsrt.GetVehicleRefForTrip(tripID); rawVehicleID != "" {
		vehicleRef = agencyID + ":VehicleRef:" + rawVehicleID
	}

	// Get complete stop sequence from GTFS static
	// Determine which key to use for static GTFS lookups
	gtfsLookupKey := tripKey
	stopSequence := c.gtfs.TripStopSeq[gtfsLookupKey]
	if len(stopSequence) == 0 {
		// Try with just trip_id if agency-prefixed key doesn't work
		gtfsLookupKey = tripID
		stopSequence = c.gtfs.TripStopSeq[gtfsLookupKey]
	}

	if len(stopSequence) == 0 {
		// Trip exists in GTFS-RT but not in GTFS static - skip it
		return nil
	}

	// Split into siri.RecordedCalls and siri.EstimatedCalls
	recordedCalls, estimatedCalls := c.buildCallSequence(tripID, gtfsLookupKey, stopSequence, now)

	// Get VehicleMode from route_type
	vehicleMode := ""
	if routeType := c.gtfs.GetRouteType(routeID); routeType > 0 {
		vehicleMode = mapGTFSRouteTypeToSIRIVehicleMode(routeType)
	}

	// Get Origin and Destination names from first/last stop in calls
	originName := ""
	destinationName := ""
	if len(stopSequence) > 0 {
		originName = c.gtfs.GetStopName(stopSequence[0])
		destinationName = c.gtfs.GetStopName(stopSequence[len(stopSequence)-1])
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
		staticArrival := gtfsTimeToUnixTimestamp(c.gtfs.GetArrivalTime(gtfsLookupKey, stopID), startDate)
		staticDeparture := gtfsTimeToUnixTimestamp(c.gtfs.GetDepartureTime(gtfsLookupKey, stopID), startDate)

		// Determine if this is a past or future stop
		isPastStop := false
		if rtDeparture > 0 && rtDeparture < now {
			isPastStop = true
		} else if rtArrival > 0 && rtArrival < now-60 { // Allow 60s grace period
			isPastStop = true
		}

		// Get stop name
		stopName := c.gtfs.GetStopName(stopID)

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
