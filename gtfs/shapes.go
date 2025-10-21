package gtfs

import (
	"math"
)

// GetStopDistanceAlongRouteForTripInMeters returns the distance in meters
func (g *GTFSIndex) GetStopDistanceAlongRouteForTripInMeters(gtfsTripKey, stopID string) float64 {
	km := g.GetStopDistanceAlongRouteForTripInKilometers(gtfsTripKey, stopID)
	if math.IsNaN(km) {
		return 0
	}
	return km * 1000
}

// GetStopDistanceAlongRouteForTripInKilometers returns the distance in kilometers
func (g *GTFSIndex) GetStopDistanceAlongRouteForTripInKilometers(gtfsTripKey, stopID string) float64 {
	stopSeq := g.TripStopSeq[gtfsTripKey]
	if len(stopSeq) == 0 {
		return 0
	}

	// Find stop index in sequence
	stopIdx := -1
	for i, s := range stopSeq {
		if s == stopID {
			stopIdx = i
			break
		}
	}
	if stopIdx < 0 {
		return 0
	}

	// Sum haversine distances from first stop to target stop
	cumKM := 0.0
	for i := 0; i < stopIdx; i++ {
		c1, ok1 := g.StopCoord[stopSeq[i]]
		c2, ok2 := g.StopCoord[stopSeq[i+1]]
		if !ok1 || !ok2 {
			continue
		}
		cumKM += HasversineKM(c1[1], c1[0], c2[1], c2[0])
	}
	return cumKM
}

// GetCoordinateAtDistanceForTrip returns a lon,lat point on the trip's shape at a target distance in KM
func (g *GTFSIndex) GetCoordinateAtDistanceForTrip(gtfsTripKey string, targetKM float64) (float64, float64, bool) {
	stopSeq := g.TripStopSeq[gtfsTripKey]
	if len(stopSeq) < 2 {
		return 0, 0, false
	}

	// Build cumulative distances between stops
	cumKM := make([]float64, len(stopSeq))
	cumKM[0] = 0
	for i := 1; i < len(stopSeq); i++ {
		c1, ok1 := g.StopCoord[stopSeq[i-1]]
		c2, ok2 := g.StopCoord[stopSeq[i]]
		if !ok1 || !ok2 {
			cumKM[i] = cumKM[i-1]
			continue
		}
		cumKM[i] = cumKM[i-1] + HasversineKM(c1[1], c1[0], c2[1], c2[0])
	}

	// Handle edge cases
	if targetKM <= 0 {
		if coord, ok := g.StopCoord[stopSeq[0]]; ok {
			return coord[0], coord[1], true
		}
		return 0, 0, false
	}
	if targetKM >= cumKM[len(cumKM)-1] {
		if coord, ok := g.StopCoord[stopSeq[len(stopSeq)-1]]; ok {
			return coord[0], coord[1], true
		}
		return 0, 0, false
	}

	// Find segment containing targetKM
	segIdx := 0
	for i := 1; i < len(cumKM); i++ {
		if cumKM[i] >= targetKM {
			segIdx = i - 1
			break
		}
	}

	// Interpolate between stops
	prevKM := cumKM[segIdx]
	nextKM := cumKM[segIdx+1]
	t := 0.0
	if nextKM > prevKM {
		t = (targetKM - prevKM) / (nextKM - prevKM)
	}

	c1, ok1 := g.StopCoord[stopSeq[segIdx]]
	c2, ok2 := g.StopCoord[stopSeq[segIdx+1]]
	if !ok1 || !ok2 {
		return 0, 0, false
	}

	lon := c1[0] + t*(c2[0]-c1[0])
	lat := c1[1] + t*(c2[1]-c1[1])

	return lon, lat, true
}

// Helpers

func HasversineKM(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	la1 := lat1 * math.Pi / 180
	la2 := lat2 * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(la1)*math.Cos(la2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
