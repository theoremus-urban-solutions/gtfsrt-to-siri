package converter

import (
	"mta/gtfsrt-to-siri/gtfsrt"
	"mta/gtfsrt-to-siri/siri"
	"mta/gtfsrt-to-siri/utils"
)

func (c *Converter) buildCall(tripID, stopID string) siri.SiriCall {
	var call siri.SiriCall
	call.VisitNumber = 1
	// timings
	if eta := c.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(tripID, stopID); eta > 0 {
		call.ExpectedArrivalTime = utils.Iso8601FromUnixSeconds(eta)
	}
	if etd := c.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(tripID, stopID); etd > 0 {
		call.ExpectedDepartureTime = utils.Iso8601FromUnixSeconds(etd)
	}
	// distances
	agency := c.Cfg.GTFS.AgencyID
	startDate := c.GTFSRT.GetStartDateForTrip(tripID)
	tripKey := gtfsrt.TripKeyForConverter(tripID, agency, startDate)
	// distance along route at the call stop
	callDistKM := c.GTFS.GetStopDistanceAlongRouteForTripInKilometers(tripKey, stopID)
	call.Extensions.Distances.CallDistanceAlongRoute = roundTo(callDistKM*1000, c.Cfg.Converter.CallDistanceAlongRouteNumDigits)
	// vehicle distance along route from snapshot
	vehKM := c.Snap.GetVehicleDistanceAlongRouteInKilometers(tripKey)
	if !isNaN(vehKM) {
		dfc := (callDistKM - vehKM) * 1000
		call.Extensions.Distances.DistanceFromCall = &dfc
		// presentable distance uses stopsFromCall unknown here; caller sets it when building lists
		call.Extensions.Distances.PresentableDistance = utils.PresentableDistance(call.Extensions.Distances.StopsFromCall, dfc/1000, 0)
	}
	return call
}

func roundTo(v float64, digits int) float64 {
	if digits < 0 {
		return v
	}
	pow := 1.0
	for i := 0; i < digits; i++ {
		pow *= 10
	}
	return float64(int(v*pow+0.5)) / pow
}

func isNaN(v float64) bool { return v != v }
