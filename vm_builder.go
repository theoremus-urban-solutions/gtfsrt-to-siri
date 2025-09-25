package gtfsrtsiri

type MonitoredVehicleJourney struct {
	LineRef                  string   `json:"LineRef"`
	DirectionRef             any      `json:"DirectionRef"`
	FramedVehicleJourneyRef  any      `json:"FramedVehicleJourneyRef"`
	JourneyPatternRef        string   `json:"JourneyPatternRef"`
	PublishedLineName        string   `json:"PublishedLineName"`
	OperatorRef              string   `json:"OperatorRef"`
	OriginRef                string   `json:"OriginRef"`
	DestinationRef           string   `json:"DestinationRef"`
	DestinationName          string   `json:"DestinationName"`
	OriginAimedDepartureTime string   `json:"OriginAimedDepartureTime"`
	SituationRef             any      `json:"SituationRef"`
	Monitored                bool     `json:"Monitored"`
	VehicleLocation          any      `json:"VehicleLocation"`
	Bearing                  *float64 `json:"Bearing"`
	ProgressRate             any      `json:"ProgressRate"`
	ProgressStatus           any      `json:"ProgressStatus"`
	VehicleRef               string   `json:"VehicleRef"`
	OnwardCalls              any      `json:"OnwardCalls"`
}

type FramedVehicleJourneyRef struct {
	DataFrameRef           string `json:"DataFrameRef"`
	DatedVehicleJourneyRef string `json:"DatedVehicleJourneyRef"`
}

type VehicleLocation struct {
	Latitude  *float64 `json:"Latitude"`
	Longitude *float64 `json:"Longitude"`
}

func (c *Converter) buildMVJ(tripID string) MonitoredVehicleJourney {
	agency := c.Cfg.GTFS.AgencyID
	startDate := c.GTFSRT.GetStartDateForTrip(tripID)
	tripKey := TripKeyForConverter(tripID, agency, startDate)

	// Prefer RT route_id; fallback to static lookup by tripKey
	routeID := c.GTFSRT.GetRouteIDForTrip(tripID)
	if routeID == "" {
		routeID = c.GTFS.GetRouteIDForTrip(tripKey)
	}
	lineRef := routeID
	if agency != "" && routeID != "" {
		lineRef = agency + "_" + routeID
	}
	direction := c.GTFSRT.GetRouteDirectionForTrip(tripID)
	if direction == "" {
		direction = c.GTFS.GetDirectionIDForTrip(tripKey)
	}
	origin := applyFieldMutators(c.GTFS.GetOriginStopIDForTrip(tripKey), c.Cfg.Converter.FieldMutators.OriginRef)
	dest := applyFieldMutators(c.GTFS.GetDestinationStopIDForTrip(tripKey), c.Cfg.Converter.FieldMutators.DestinationRef)
	head := c.GTFS.GetTripHeadsign(tripKey)
	pub := c.GTFS.GetRouteShortName(routeID)
	vehRef := c.GTFSRT.GetVehicleRefForTrip(tripID)
	if agency != "" && vehRef != "" {
		vehRef = agency + "_" + vehRef
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

	// FramedVehicleJourneyRef
	dataFrameRef := ""
	if len(startDate) == 8 { // YYYYMMDD
		dataFrameRef = startDate[:4] + "-" + startDate[4:6] + "-" + startDate[6:8]
	}
	dvj := c.CfGDatedVehicleJourneyRef(tripKey, agency)

	// OriginAimedDepartureTime fallback order: RT dep at origin, else RT arr at origin, else empty (scheduled not available yet)
	originAimed := ""
	if origin != "" {
		if dep := c.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(tripID, origin); dep > 0 {
			originAimed = iso8601FromUnixSeconds(dep)
		} else if arr := c.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(tripID, origin); arr > 0 {
			originAimed = iso8601FromUnixSeconds(arr)
		}
	}

	// OnwardCalls: built elsewhere with limits
	onward := c.buildOnwardCalls(tripID, -1, "", false)

	return MonitoredVehicleJourney{
		LineRef:                 lineRef,
		DirectionRef:            direction,
		FramedVehicleJourneyRef: FramedVehicleJourneyRef{DataFrameRef: dataFrameRef, DatedVehicleJourneyRef: dvj},
		JourneyPatternRef: func() string {
			sh := c.GTFS.GetShapeIDForTrip(tripKey)
			if sh == "" {
				return ""
			}
			if agency != "" {
				return agency + "_" + sh
			}
			return sh
		}(),
		PublishedLineName:        pub,
		OperatorRef:              agency,
		OriginRef:                origin,
		DestinationRef:           dest,
		DestinationName:          head,
		OriginAimedDepartureTime: originAimed,
		SituationRef:             nil,
		Monitored:                true,
		VehicleLocation:          VehicleLocation{Latitude: latPtr, Longitude: lonPtr},
		Bearing:                  bearing,
		ProgressRate:             nil,
		ProgressStatus:           nil,
		VehicleRef:               vehRef,
		OnwardCalls:              onward,
	}
}

// CfGDatedVehicleJourneyRef returns a DatedVehicleJourneyRef; concat agency + full trip id based on strategy
func (c *Converter) CfGDatedVehicleJourneyRef(tripKey, agency string) string {
	if agency != "" {
		return agency + "_" + c.GTFS.GetFullTripIDForTrip(tripKey)
	}
	return c.GTFS.GetFullTripIDForTrip(tripKey)
}

// buildOnwardCalls builds OnwardCalls with optional max limit and behavior for StopMonitoring
func (c *Converter) buildOnwardCalls(tripID string, maxOnward int, selectedStopID string, stopMonitoring bool) any {
	stops := c.GTFSRT.GetOnwardStopIDsForTrip(tripID)
	if len(stops) == 0 {
		return nil
	}
	limit := len(stops)
	if maxOnward >= 0 && maxOnward < limit {
		limit = maxOnward
	}
	if stopMonitoring && selectedStopID != "" {
		idx := c.GTFSRT.GetIndexOfStopInStopTimeUpdatesForTrip(tripID, selectedStopID)
		if idx >= 0 && (idx+1) > limit {
			limit = idx + 1
		}
	}
	// Base distances
	agency := c.Cfg.GTFS.AgencyID
	startDate := c.GTFSRT.GetStartDateForTrip(tripID)
	tripKey := TripKeyForConverter(tripID, agency, startDate)

	calls := make([]SiriCall, 0, limit)
	// compute vehicle distance and next-stop distance for presentable distance tuning
	vehKMOverall := c.Snap.GetVehicleDistanceAlongRouteInKilometers(tripKey)
	nextStopDistKM := 0.0
	if len(stops) > 0 && !isNaN(vehKMOverall) {
		nextStopDistKM = c.GTFS.GetStopDistanceAlongRouteForTripInKilometers(tripKey, stops[0]) - vehKMOverall
		if nextStopDistKM < 0 {
			nextStopDistKM = 0
		}
	}
	// fill calls and distances
	for i := 0; i < limit && i < len(stops); i++ {
		sid := stops[i]
		call := c.buildCall(tripID, sid)
		call.StopPointRef = applyFieldMutators(sid, c.Cfg.Converter.FieldMutators.StopPointRef)
		call.StopPointName = c.GTFS.GetStopName(sid)
		if eta := c.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(tripID, sid); eta > 0 {
			call.ExpectedArrivalTime = iso8601FromUnixSeconds(eta)
		}
		if etd := c.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(tripID, sid); etd > 0 {
			call.ExpectedDepartureTime = iso8601FromUnixSeconds(etd)
		}
		// Distances along route
		callDistKM := c.GTFS.GetStopDistanceAlongRouteForTripInKilometers(tripKey, sid)
		call.Extensions.Distances.StopsFromCall = i
		// per config rounding
		call.Extensions.Distances.CallDistanceAlongRoute = roundTo(callDistKM*1000, c.Cfg.Converter.CallDistanceAlongRouteNumDigits)
		// vehicle position distance from snapshot
		vehKM := vehKMOverall
		if !isNaN(vehKM) {
			dfc := (callDistKM - vehKM) * 1000
			call.Extensions.Distances.DistanceFromCall = &dfc
			// presentable distance: use distance to current call and distance to immediate next stop
			call.Extensions.Distances.PresentableDistance = presentableDistance(i, callDistKM-vehKM, nextStopDistKM)
		}
		calls = append(calls, call)
	}
	return map[string]any{"OnwardCall": calls}
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
