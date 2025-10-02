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
	shapeID := g.GetShapeIDForTrip(gtfsTripKey)
	if shapeID == "" {
		return 0
	}
	pts := g.ShapePoints[shapeID]
	if len(pts) < 2 {
		return 0
	}
	coord, ok := g.stopCoord[stopID]
	if !ok {
		return 0
	}
	segIdx, t, _ := NearestSegmentProjection(pts, coord)
	cum := g.ShapeCumKM[shapeID]
	if segIdx < 0 || segIdx >= len(cum) {
		return 0
	}
	if segIdx == len(pts)-1 {
		return cum[segIdx]
	}
	// add fractional distance within the segment
	segKM := HasversineKM(pts[segIdx][1], pts[segIdx][0], pts[segIdx+1][1], pts[segIdx+1][0])
	return cum[segIdx] + t*segKM
}

// GetCoordinateAtDistanceForTrip returns a lon,lat point on the trip's shape at a target distance in KM
func (g *GTFSIndex) GetCoordinateAtDistanceForTrip(gtfsTripKey string, targetKM float64) (float64, float64, bool) {
	shapeID := g.GetShapeIDForTrip(gtfsTripKey)
	pts := g.ShapePoints[shapeID]
	cum := g.ShapeCumKM[shapeID]
	if len(pts) == 0 || len(cum) == 0 {
		return 0, 0, false
	}
	if targetKM <= 0 {
		return pts[0][0], pts[0][1], true
	}
	if targetKM >= cum[len(cum)-1] {
		last := pts[len(pts)-1]
		return last[0], last[1], true
	}
	// binary search for segment
	lo := 0
	hi := len(cum) - 1
	for lo < hi {
		mid := (lo + hi) / 2
		if cum[mid] < targetKM {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	i := lo
	if i == 0 {
		i = 1
	}
	prevKM := cum[i-1]
	nextKM := cum[i]
	segLen := nextKM - prevKM
	t := 0.0
	if segLen > 0 {
		t = (targetKM - prevKM) / segLen
	}
	ax, ay := pts[i-1][0], pts[i-1][1]
	bx, by := pts[i][0], pts[i][1]
	lon := ax + t*(bx-ax)
	lat := ay + t*(by-ay)
	return lon, lat, true
}

// Helpers

func cumulativeKM(pts [][2]float64) []float64 {
	cum := make([]float64, len(pts))
	if len(pts) == 0 {
		return cum
	}
	cum[0] = 0
	for i := 1; i < len(pts); i++ {
		cum[i] = cum[i-1] + HasversineKM(pts[i-1][1], pts[i-1][0], pts[i][1], pts[i][0])
	}
	return cum
}

func nearestPointIndex(pts [][2]float64, coord [2]float64) int {
	best := -1
	bestD := math.MaxFloat64
	for i, p := range pts {
		d := HasversineKM(coord[1], coord[0], p[1], p[0])
		if d < bestD {
			bestD = d
			best = i
		}
	}
	return best
}

// nearestSegmentProjection finds the segment index i (between pts[i] and pts[i+1])
// that is closest to the given coordinate, and returns the clamped projection
// parameter t in [0,1] along that segment and the snapped lon/lat point.
func NearestSegmentProjection(pts [][2]float64, coord [2]float64) (int, float64, [2]float64) {
	bestIdx := -1
	bestT := 0.0
	var bestSnap [2]float64
	bestDist2 := math.MaxFloat64
	cx := coord[0]
	cy := coord[1]
	for i := 0; i+1 < len(pts); i++ {
		ax := pts[i][0]
		ay := pts[i][1]
		bx := pts[i+1][0]
		by := pts[i+1][1]
		vx := bx - ax
		vy := by - ay
		wx := cx - ax
		wy := cy - ay
		denom := vx*vx + vy*vy
		t := 0.0
		if denom > 0 {
			t = (wx*vx + wy*vy) / denom
		}
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
		sx := ax + t*vx
		sy := ay + t*vy
		dx := cx - sx
		dy := cy - sy
		dist2 := dx*dx + dy*dy
		if dist2 < bestDist2 {
			bestDist2 = dist2
			bestIdx = i
			bestT = t
			bestSnap = [2]float64{sx, sy}
		}
	}
	return bestIdx, bestT, bestSnap
}

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
