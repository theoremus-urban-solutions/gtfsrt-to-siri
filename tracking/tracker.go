package tracking

import (
	"math"
	"time"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfs"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
)

type Snapshot struct {
	gtfsrtTimestamp int64
	previous        *Snapshot
	trainLocations  map[string]*TrainLocation
	tripKeyToGTFS   map[string]*gtfs.GTFSIndex
}

type TrainLocation struct {
	State                 TrainState
	LocationGeoJSONType   string
	Coordinates           [][]float64
	Bearing               float64
	StartDistAlongRouteKM float64
	LineDistanceKM        float64
	ImmediateStopID       string
}

type TrainState struct {
	FirstOccurrence       bool
	KnewLocation          bool
	AtStop                bool
	AtOrigin              bool
	AtDestination         bool
	NoETA                 bool
	BadPreviousETA        bool
	OutOfSequenceStops    bool
	AtIntermediateStop    bool
	HasMoved              bool
	SameImmediateNextStop bool
	NoStopTimeUpdate      bool
}

var previousSnapshot *Snapshot

func NewSnapshot(gtfsIdx *gtfs.GTFSIndex, rt *gtfsrt.GTFSRTWrapper, agencyID string) *Snapshot {
	ts := rt.GetTimestampForFeedMessage()
	if previousSnapshot != nil && ts < previousSnapshot.gtfsrtTimestamp {
		return previousSnapshot
	}
	if previousSnapshot != nil && ts == previousSnapshot.gtfsrtTimestamp {
		return previousSnapshot
	}
	s := &Snapshot{
		gtfsrtTimestamp: ts,
		previous:        previousSnapshot,
		trainLocations:  map[string]*TrainLocation{},
		tripKeyToGTFS:   map[string]*gtfs.GTFSIndex{},
	}
	// Fill using RT data; interpolate between stops when possible
	agency := agencyID
	for _, rtTrip := range rt.GetAllMonitoredTrips() {
		startDate := rt.GetStartDateForTrip(rtTrip)
		tripKey := gtfsrt.TripKeyForConverter(rtTrip, agency, startDate)
		// If GTFS-RT vehicle location exists, use it; else approximate between origin and next stops
		var coords [][]float64
		var bearing float64
		if lat, ok := rt.GetVehicleLatForTrip(rtTrip); ok {
			if lon, ok2 := rt.GetVehicleLonForTrip(rtTrip); ok2 {
				coords = [][]float64{{lon, lat}}
			}
		}
		if b, ok := rt.GetVehicleBearingForTrip(rtTrip); ok {
			bearing = b
		} else {
			bearing = math.NaN()
		}
		// Interpolation fallback
		startDistKM := 0.0
		if len(coords) == 0 {
			onward := rt.GetOnwardStopIDsForTrip(rtTrip)
			if len(onward) > 0 {
				// If we have times for the next two stops, interpolate by ETA
				now := nowEpoch()
				s0 := onward[0]
				eta0 := rt.GetExpectedArrivalTimeAtStopForTrip(rtTrip, s0)
				// Try to find a second stop to form a segment in distance space
				var s1 string
				if len(onward) > 1 {
					s1 = onward[1]
				}
				// Compute distances along route
				d0 := gtfsIdx.GetStopDistanceAlongRouteForTripInKilometers(tripKey, s0)
				var d1 float64
				if s1 != "" {
					d1 = gtfsIdx.GetStopDistanceAlongRouteForTripInKilometers(tripKey, s1)
				} else {
					d1 = d0
				}
				// Interpolate current distance between d0 and d1 based on time
				curKM := d0
				if s1 != "" {
					eta1 := rt.GetExpectedArrivalTimeAtStopForTrip(rtTrip, s1)
					if eta0 > 0 && eta1 > eta0 {
						if now <= eta0 {
							curKM = d0
						} else if now >= eta1 {
							curKM = d1
						} else {
							frac := float64(now-eta0) / float64(eta1-eta0)
							curKM = d0 + frac*(d1-d0)
						}
					}
				}
				// Map distance to coordinate on shape
				lon, lat, ok := gtfsIdx.GetCoordinateAtDistanceForTrip(tripKey, curKM)
				if ok {
					coords = [][]float64{{lon, lat}}
				}
			}
		} else {
			// derive distance from RT position by projection onto stop segments
			stopSeq := gtfsIdx.TripStopSeq[tripKey]
			if len(stopSeq) >= 2 {
				vehCoord := [2]float64{coords[0][0], coords[0][1]}

				minDist := math.MaxFloat64
				bestSegIdx := 0
				bestT := 0.0

				for i := 0; i < len(stopSeq)-1; i++ {
					c1, ok1 := gtfsIdx.StopCoord[stopSeq[i]]
					c2, ok2 := gtfsIdx.StopCoord[stopSeq[i+1]]
					if !ok1 || !ok2 {
						continue
					}

					// Project vehicle onto segment between c1 and c2
					vx := c2[0] - c1[0]
					vy := c2[1] - c1[1]
					wx := vehCoord[0] - c1[0]
					wy := vehCoord[1] - c1[1]

					denom := vx*vx + vy*vy
					t := 0.0
					if denom > 0 {
						t = (wx*vx + wy*vy) / denom
						if t < 0 {
							t = 0
						} else if t > 1 {
							t = 1
						}
					}

					px := c1[0] + t*vx
					py := c1[1] + t*vy

					dx := vehCoord[0] - px
					dy := vehCoord[1] - py
					dist := dx*dx + dy*dy

					if dist < minDist {
						minDist = dist
						bestSegIdx = i
						bestT = t
					}
				}

				// Compute cumulative distance to best segment + fractional
				startDistKM = 0.0
				for j := 0; j < bestSegIdx; j++ {
					sc1, ok1 := gtfsIdx.StopCoord[stopSeq[j]]
					sc2, ok2 := gtfsIdx.StopCoord[stopSeq[j+1]]
					if ok1 && ok2 {
						startDistKM += gtfs.HasversineKM(sc1[1], sc1[0], sc2[1], sc2[0])
					}
				}
				// Add fractional distance within best segment
				if bestSegIdx < len(stopSeq)-1 {
					c1, ok1 := gtfsIdx.StopCoord[stopSeq[bestSegIdx]]
					c2, ok2 := gtfsIdx.StopCoord[stopSeq[bestSegIdx+1]]
					if ok1 && ok2 {
						segKM := gtfs.HasversineKM(c1[1], c1[0], c2[1], c2[0])
						startDistKM += bestT * segKM
					}
				}
			}
		}
		s.trainLocations[tripKey] = &TrainLocation{
			State:                 TrainState{},
			LocationGeoJSONType:   "Point",
			Coordinates:           coords,
			Bearing:               bearing,
			StartDistAlongRouteKM: startDistKM,
			LineDistanceKM:        math.NaN(),
			ImmediateStopID:       "",
		}
		s.tripKeyToGTFS[tripKey] = gtfsIdx
	}
	previousSnapshot = s
	return s
}

func (s *Snapshot) GetLatitude(gtfsTripKey string) *float64 {
	loc := s.trainLocations[gtfsTripKey]
	if loc == nil || len(loc.Coordinates) == 0 || len(loc.Coordinates[0]) < 2 {
		return nil
	}
	lat := loc.Coordinates[0][1]
	if math.IsNaN(lat) {
		return nil
	}
	return &lat
}

func (s *Snapshot) GetLongitude(gtfsTripKey string) *float64 {
	loc := s.trainLocations[gtfsTripKey]
	if loc == nil || len(loc.Coordinates) == 0 || len(loc.Coordinates[0]) < 2 {
		return nil
	}
	lon := loc.Coordinates[0][0]
	if math.IsNaN(lon) {
		return nil
	}
	return &lon
}

func (s *Snapshot) GetBearing(gtfsTripKey string) *float64 {
	loc := s.trainLocations[gtfsTripKey]
	if loc == nil || math.IsNaN(loc.Bearing) {
		return nil
	}
	b := loc.Bearing
	return &b
}

func (s *Snapshot) GetVehicleDistanceAlongRouteInKilometers(gtfsTripKey string) float64 {
	loc := s.trainLocations[gtfsTripKey]
	if loc == nil {
		return math.NaN()
	}
	return loc.StartDistAlongRouteKM
}

// placeholder used later for health
func (s *Snapshot) GetTimestamp() int64 { return s.gtfsrtTimestamp }

// helper until RT feed is wired
func nowEpoch() int64 { return time.Now().Unix() }
