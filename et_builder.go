package gtfsrtsiri

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
	DirectionRef            string                  `json:"DirectionRef"`
	FramedVehicleJourneyRef FramedVehicleJourneyRef `json:"FramedVehicleJourneyRef"`
	VehicleMode             string                  `json:"VehicleMode,omitempty"`
	OriginName              string                  `json:"OriginName,omitempty"`
	DestinationName         string                  `json:"DestinationName,omitempty"`
	Monitored               bool                    `json:"Monitored"`
	DataSource              string                  `json:"DataSource,omitempty"`
	VehicleRef              string                  `json:"VehicleRef,omitempty"`
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
	StopPointRef        string `json:"StopPointRef"`
	Order               int    `json:"Order"`
	StopPointName       string `json:"StopPointName,omitempty"`
	Cancellation        bool   `json:"Cancellation,omitempty"`
	AimedArrivalTime    string `json:"AimedArrivalTime,omitempty"`
	ActualArrivalTime   string `json:"ActualArrivalTime,omitempty"`
	AimedDepartureTime  string `json:"AimedDepartureTime,omitempty"`
	ActualDepartureTime string `json:"ActualDepartureTime,omitempty"`
}

type EstimatedCall struct {
	StopPointRef          string `json:"StopPointRef"`
	Order                 int    `json:"Order"`
	StopPointName         string `json:"StopPointName,omitempty"`
	Cancellation          bool   `json:"Cancellation,omitempty"`
	AimedArrivalTime      string `json:"AimedArrivalTime,omitempty"`
	ExpectedArrivalTime   string `json:"ExpectedArrivalTime,omitempty"`
	AimedDepartureTime    string `json:"AimedDepartureTime,omitempty"`
	ExpectedDepartureTime string `json:"ExpectedDepartureTime,omitempty"`
	ArrivalStatus         string `json:"ArrivalStatus,omitempty"`
	DepartureStatus       string `json:"DepartureStatus,omitempty"`
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

	// Get vehicle ref if available
	vehicleRef := c.GTFSRT.GetVehicleRefForTrip(tripID)

	// Get complete stop sequence from GTFS static
	stopSequence := c.GTFS.tripStopSeq[tripKey]
	if len(stopSequence) == 0 {
		// Try with just trip_id if agency-prefixed key doesn't work
		stopSequence = c.GTFS.tripStopSeq[tripID]
	}

	if len(stopSequence) == 0 {
		return nil
	}

	// Split into RecordedCalls and EstimatedCalls
	recordedCalls, estimatedCalls := c.buildCallSequence(tripID, tripKey, stopSequence, now)

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
		VehicleRef:             vehicleRef,
		OperatorRef:            operatorRef,
		RecordedCalls:          recordedCalls,
		EstimatedCalls:         estimatedCalls,
		IsCompleteStopSequence: true,
	}

	return journey
}

func (c *Converter) buildCallSequence(tripID, tripKey string, stopSequence []string, now int64) ([]RecordedCall, []EstimatedCall) {
	recordedCalls := []RecordedCall{}
	estimatedCalls := []EstimatedCall{}

	for order, stopID := range stopSequence {
		// Get real-time arrival/departure times
		rtArrival := c.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(tripID, stopID)
		rtDeparture := c.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(tripID, stopID)

		// Determine if this is a past or future stop
		isPastStop := false
		if rtDeparture > 0 && rtDeparture < now {
			isPastStop = true
		} else if rtArrival > 0 && rtArrival < now-60 { // Allow 60s grace period
			isPastStop = true
		}

		// Get stop name
		stopName := c.GTFS.GetStopName(stopID)

		// Apply field mutators to stop ref
		stopPointRef := applyFieldMutators(stopID, c.Cfg.Converter.FieldMutators.StopPointRef)

		if isPastStop {
			// RecordedCall
			call := RecordedCall{
				StopPointRef:  stopPointRef,
				Order:         order + 1,
				StopPointName: stopName,
				Cancellation:  false,
			}

			// Set actual times from GTFS-RT
			if rtArrival > 0 {
				call.ActualArrivalTime = iso8601ExtendedFromUnixSeconds(rtArrival)
				// Use same as aimed for now (we don't have static schedule times easily accessible)
				call.AimedArrivalTime = iso8601ExtendedFromUnixSeconds(rtArrival)
			}
			if rtDeparture > 0 {
				call.ActualDepartureTime = iso8601ExtendedFromUnixSeconds(rtDeparture)
				call.AimedDepartureTime = iso8601ExtendedFromUnixSeconds(rtDeparture)
			}

			recordedCalls = append(recordedCalls, call)
		} else {
			// EstimatedCall
			call := EstimatedCall{
				StopPointRef:  stopPointRef,
				Order:         order + 1,
				StopPointName: stopName,
				Cancellation:  false,
			}

			// Set expected times from GTFS-RT
			if rtArrival > 0 {
				call.ExpectedArrivalTime = iso8601ExtendedFromUnixSeconds(rtArrival)
				// Use current time or expected as aimed (simplified)
				call.AimedArrivalTime = iso8601ExtendedFromUnixSeconds(rtArrival)
				call.ArrivalStatus = calculateStatus(rtArrival, rtArrival)
			}
			if rtDeparture > 0 {
				call.ExpectedDepartureTime = iso8601ExtendedFromUnixSeconds(rtDeparture)
				call.AimedDepartureTime = iso8601ExtendedFromUnixSeconds(rtDeparture)
				call.DepartureStatus = calculateStatus(rtDeparture, rtDeparture)
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
