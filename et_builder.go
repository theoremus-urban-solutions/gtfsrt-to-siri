package gtfsrtsiri

import (
	"fmt"
	"time"
)

// EstimatedTimetable delivery types
type EstimatedTimetable struct {
	ResponseTimestamp            string                         `json:"ResponseTimestamp"`
	EstimatedJourneyVersionFrame []EstimatedJourneyVersionFrame `json:"EstimatedJourneyVersionFrame"`
}

type EstimatedJourneyVersionFrame struct {
	RecordedAtTime          string                    `json:"RecordedAtTime"`
	EstimatedVehicleJourney []EstimatedVehicleJourney `json:"EstimatedVehicleJourney"`
}

type EstimatedVehicleJourney struct {
	RecordedAtTime          string                  `json:"RecordedAtTime"`
	LineRef                 string                  `json:"LineRef"`
	VehicleRef              string                  `json:"VehicleRef,omitempty"`
	DirectionRef            string                  `json:"DirectionRef"`
	FramedVehicleJourneyRef FramedVehicleJourneyRef `json:"FramedVehicleJourneyRef"`
	VehicleMode             string                  `json:"VehicleMode,omitempty"`
	OriginName              string                  `json:"OriginName,omitempty"`
	DestinationName         string                  `json:"DestinationName,omitempty"`
	Monitored               bool                    `json:"Monitored"`
	DataSource              string                  `json:"DataSource,omitempty"`
	OperatorRef             string                  `json:"OperatorRef,omitempty"`
	RecordedCalls           []RecordedCall          `json:"RecordedCalls,omitempty"`
	EstimatedCalls          []EstimatedCall         `json:"EstimatedCalls,omitempty"`
	IsCompleteStopSequence  bool                    `json:"IsCompleteStopSequence"`
}

type FramedVehicleJourneyRef struct {
	DataFrameRef           string `json:"DataFrameRef"`
	DatedVehicleJourneyRef string `json:"DatedVehicleJourneyRef"`
}

type RecordedCall struct {
	StopPointRef        string `json:"StopPointRef" xml:"StopPointRef"`
	Order               int    `json:"Order" xml:"Order"`
	StopPointName       string `json:"StopPointName,omitempty" xml:"StopPointName,omitempty"`
	Cancellation        bool   `json:"Cancellation,omitempty" xml:"Cancellation"`
	RequestStop         bool   `json:"RequestStop,omitempty" xml:"RequestStop"`
	AimedArrivalTime    string `json:"AimedArrivalTime,omitempty" xml:"AimedArrivalTime,omitempty"`
	ActualArrivalTime   string `json:"ActualArrivalTime,omitempty" xml:"ActualArrivalTime,omitempty"`
	AimedDepartureTime  string `json:"AimedDepartureTime,omitempty" xml:"AimedDepartureTime,omitempty"`
	ActualDepartureTime string `json:"ActualDepartureTime,omitempty" xml:"ActualDepartureTime,omitempty"`
}

type EstimatedCall struct {
	StopPointRef          string `json:"StopPointRef" xml:"StopPointRef"`
	Order                 int    `json:"Order" xml:"Order"`
	StopPointName         string `json:"StopPointName,omitempty" xml:"StopPointName,omitempty"`
	Cancellation          bool   `json:"Cancellation,omitempty" xml:"Cancellation"`
	RequestStop           bool   `json:"RequestStop,omitempty" xml:"RequestStop"`
	AimedArrivalTime      string `json:"AimedArrivalTime,omitempty" xml:"AimedArrivalTime,omitempty"`
	ExpectedArrivalTime   string `json:"ExpectedArrivalTime,omitempty" xml:"ExpectedArrivalTime,omitempty"`
	AimedDepartureTime    string `json:"AimedDepartureTime,omitempty" xml:"AimedDepartureTime,omitempty"`
	ExpectedDepartureTime string `json:"ExpectedDepartureTime,omitempty" xml:"ExpectedDepartureTime,omitempty"`
	ArrivalStatus         string `json:"ArrivalStatus,omitempty" xml:"ArrivalStatus,omitempty"`
	DepartureStatus       string `json:"DepartureStatus,omitempty" xml:"DepartureStatus,omitempty"`
}

// BuildEstimatedTimetable converts GTFS-RT data to SIRI ET format
func (c *Converter) BuildEstimatedTimetable() EstimatedTimetable {
	timestamp := c.GTFSRT.GetTimestampForFeedMessage()
	now := timestamp
	agencyID := c.Cfg.GTFS.AgencyID
	if agencyID == "" {
		agencyID = "UNKNOWN"
	}

	// Get all active trips from GTFS-RT
	allTrips := []string{}
	for tripID := range c.GTFSRT.trips {
		allTrips = append(allTrips, tripID)
	}
	journeys := make([]EstimatedVehicleJourney, 0, len(allTrips))

	for _, tripID := range allTrips {
		journey := c.buildEstimatedVehicleJourney(tripID, now, agencyID)
		if journey != nil {
			journeys = append(journeys, *journey)
		}
	}

	frame := EstimatedJourneyVersionFrame{
		RecordedAtTime:          iso8601ExtendedFromUnixSeconds(timestamp),
		EstimatedVehicleJourney: journeys,
	}

	return EstimatedTimetable{
		ResponseTimestamp:            iso8601ExtendedFromUnixSeconds(timestamp),
		EstimatedJourneyVersionFrame: []EstimatedJourneyVersionFrame{frame},
	}
}

func (c *Converter) buildEstimatedVehicleJourney(tripID string, now int64, agencyID string) *EstimatedVehicleJourney {
	// Get route and direction
	routeID := c.GTFSRT.GetRouteIDForTrip(tripID)
	if routeID == "" {
		return nil
	}

	directionID := c.GTFSRT.GetRouteDirectionForTrip(tripID)
	if directionID == "" {
		directionID = "0"
	}

	// Get trip key
	startDate := c.GTFSRT.GetStartDateForTrip(tripID)
	tripKey := TripKeyForConverter(tripID, agencyID, startDate)

	// Build FramedVehicleJourneyRef
	dataFrameRef := startDate
	if dataFrameRef == "" {
		dataFrameRef = iso8601DateFromUnixSeconds(now)
	}
	datedVehicleJourneyRef := agencyID + ":ServiceJourney:" + tripID

	// Get vehicle ref if available and format as {codespace}:VehicleRef:{vehicle_id}
	vehicleRef := ""
	if rawVehicleID := c.GTFSRT.GetVehicleRefForTrip(tripID); rawVehicleID != "" {
		vehicleRef = agencyID + ":VehicleRef:" + rawVehicleID
	}

	// Get complete stop sequence from GTFS static
	// Determine which key to use for static GTFS lookups
	gtfsLookupKey := tripKey
	stopSequence := c.GTFS.tripStopSeq[gtfsLookupKey]
	if len(stopSequence) == 0 {
		// Try with just trip_id if agency-prefixed key doesn't work
		gtfsLookupKey = tripID
		stopSequence = c.GTFS.tripStopSeq[gtfsLookupKey]
	}

	if len(stopSequence) == 0 {
		return nil
	}

	// Split into RecordedCalls and EstimatedCalls
	recordedCalls, estimatedCalls := c.buildCallSequence(tripID, gtfsLookupKey, stopSequence, now)

	// Get VehicleMode from route_type
	vehicleMode := ""
	if routeType := c.GTFS.GetRouteType(routeID); routeType > 0 {
		vehicleMode = mapGTFSRouteTypeToSIRIVehicleMode(routeType)
	}

	// Get Origin and Destination names from first/last stop in calls
	originName := ""
	destinationName := ""
	if len(stopSequence) > 0 {
		originName = c.GTFS.GetStopName(stopSequence[0])
		destinationName = c.GTFS.GetStopName(stopSequence[len(stopSequence)-1])
	}

	// OperatorRef from agency_name
	operatorRef := agencyID
	if agencyName := c.GTFS.GetAgencyName(); agencyName != "" {
		operatorRef = agencyID + ":Operator:" + agencyName
	}

	// Monitored: true if trip is currently ongoing (has both past and future stops)
	monitored := len(recordedCalls) > 0 && len(estimatedCalls) > 0

	journey := &EstimatedVehicleJourney{
		RecordedAtTime: iso8601ExtendedFromUnixSeconds(now),
		LineRef:        agencyID + ":Line:" + routeID,
		VehicleRef:     vehicleRef,
		DirectionRef:   directionID,
		FramedVehicleJourneyRef: FramedVehicleJourneyRef{
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

func (c *Converter) buildCallSequence(tripID, gtfsLookupKey string, stopSequence []string, now int64) ([]RecordedCall, []EstimatedCall) {
	recordedCalls := []RecordedCall{}
	estimatedCalls := []EstimatedCall{}
	agencyID := c.Cfg.GTFS.AgencyID
	if agencyID == "" {
		agencyID = "UNKNOWN"
	}

	// Get start_date for time conversion
	startDate := c.GTFSRT.GetStartDateForTrip(tripID)
	// If no start_date from GTFS-RT, use today's date
	if startDate == "" {
		startDate = time.Unix(now, 0).Format("20060102")
	}

	for order, stopID := range stopSequence {
		// Get real-time arrival/departure times
		rtArrival := c.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(tripID, stopID)
		rtDeparture := c.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(tripID, stopID)

		// Get static GTFS times using gtfsLookupKey (the key that successfully found stopSequence)
		staticArrival := gtfsTimeToUnixTimestamp(c.GTFS.GetArrivalTime(gtfsLookupKey, stopID), startDate)
		staticDeparture := gtfsTimeToUnixTimestamp(c.GTFS.GetDepartureTime(gtfsLookupKey, stopID), startDate)

		// Determine if this is a past or future stop
		isPastStop := false
		if rtDeparture > 0 && rtDeparture < now {
			isPastStop = true
		} else if rtArrival > 0 && rtArrival < now-60 { // Allow 60s grace period
			isPastStop = true
		}

		// Get stop name
		stopName := c.GTFS.GetStopName(stopID)

		// Format StopPointRef as {codespace}:Quay:{stop_id}, then apply field mutators
		stopPointRef := applyFieldMutators(agencyID+":Quay:"+stopID, c.Cfg.Converter.FieldMutators.StopPointRef)

		// Check if cancelled (schedule_relationship = 1 SKIPPED)
		schedRel := c.GTFSRT.GetScheduleRelationshipForStop(tripID, stopID)
		isCancelled := schedRel == 1

		// Check if request stop (pickup_type or drop_off_type = 2 or 3) using gtfsLookupKey
		pickupType := c.GTFS.GetPickupType(gtfsLookupKey, stopID)
		dropOffType := c.GTFS.GetDropOffType(gtfsLookupKey, stopID)
		isRequestStop := pickupType == 2 || pickupType == 3 || dropOffType == 2 || dropOffType == 3

		if isPastStop {
			// RecordedCall
			call := RecordedCall{
				StopPointRef:  stopPointRef,
				Order:         order + 1,
				StopPointName: stopName,
				Cancellation:  isCancelled,
				RequestStop:   isRequestStop,
			}

			// Set aimed times from static GTFS
			if staticArrival > 0 {
				call.AimedArrivalTime = iso8601ExtendedFromUnixSeconds(staticArrival)
			}
			if staticDeparture > 0 {
				call.AimedDepartureTime = iso8601ExtendedFromUnixSeconds(staticDeparture)
			}

			// Set actual times from GTFS-RT
			if rtArrival > 0 {
				call.ActualArrivalTime = iso8601ExtendedFromUnixSeconds(rtArrival)
			}
			if rtDeparture > 0 {
				call.ActualDepartureTime = iso8601ExtendedFromUnixSeconds(rtDeparture)
			}

			recordedCalls = append(recordedCalls, call)
		} else {
			// EstimatedCall
			call := EstimatedCall{
				StopPointRef:  stopPointRef,
				Order:         order + 1,
				StopPointName: stopName,
				Cancellation:  isCancelled,
				RequestStop:   isRequestStop,
			}

			// Set aimed times from static GTFS
			if staticArrival > 0 {
				call.AimedArrivalTime = iso8601ExtendedFromUnixSeconds(staticArrival)
			}
			if staticDeparture > 0 {
				call.AimedDepartureTime = iso8601ExtendedFromUnixSeconds(staticDeparture)
			}

			// Set expected times from GTFS-RT and calculate status
			if rtArrival > 0 {
				call.ExpectedArrivalTime = iso8601ExtendedFromUnixSeconds(rtArrival)
				if staticArrival > 0 {
					call.ArrivalStatus = calculateStatus(rtArrival, staticArrival)
				} else {
					call.ArrivalStatus = "onTime" // Default if no static time
				}
			}
			if rtDeparture > 0 {
				call.ExpectedDepartureTime = iso8601ExtendedFromUnixSeconds(rtDeparture)
				if staticDeparture > 0 {
					call.DepartureStatus = calculateStatus(rtDeparture, staticDeparture)
				} else {
					call.DepartureStatus = "onTime" // Default if no static time
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
