package gtfs

import (
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/config"
)

// GTFSIndex stores GTFS static data in memory for fast lookups
type GTFSIndex struct {
	agencyID        string
	agencyTZ        string
	agencyName      string                    // agency_name from agency.txt
	routeShortNames map[string]string         // route_id -> short_name
	routeTypes      map[string]int            // route_id -> route_type (GTFS enum)
	routes          map[string]struct{}       // route existence set
	tripToRoute     map[string]string         // trip_id -> route_id
	tripHeadsign    map[string]string         // trip_id -> headsign
	tripOriginStop  map[string]string         // trip_id -> first stop_id
	tripDestStop    map[string]string         // trip_id -> last stop_id
	tripDirection   map[string]string         // trip_id -> direction_id ("0"|"1")
	tripShapeID     map[string]string         // trip_id -> shape_id
	tripBlockID     map[string]string         // trip_id -> block_id
	TripStopSeq     map[string][]string       // trip_id -> ordered stop_ids
	tripStopIdx     map[string]map[string]int // trip_id -> stop_id -> index
	stopNames       map[string]string         // stop_id -> name
	stopCoord       map[string][2]float64     // stop_id -> [lon,lat]
	ShapePoints     map[string][][2]float64   // shape_id -> ordered points [lon,lat]
	ShapeCumKM      map[string][]float64      // shape_id -> cumulative km at each point
	// New fields for ET support
	stopTimePickupType  map[string]map[string]int    // trip_id -> stop_id -> pickup_type
	stopTimeDropOffType map[string]map[string]int    // trip_id -> stop_id -> drop_off_type
	stopTimeArrival     map[string]map[string]string // trip_id -> stop_id -> arrival_time (HH:MM:SS)
	stopTimeDeparture   map[string]map[string]string // trip_id -> stop_id -> departure_time (HH:MM:SS)
}

// NewGTFSIndex creates a new empty GTFS index
func NewGTFSIndex(indexedSchedulePath, indexedSpatialPath string) (*GTFSIndex, error) {
	return &GTFSIndex{
		routeShortNames:     map[string]string{},
		routeTypes:          map[string]int{},
		routes:              map[string]struct{}{},
		tripToRoute:         map[string]string{},
		tripHeadsign:        map[string]string{},
		tripOriginStop:      map[string]string{},
		tripDestStop:        map[string]string{},
		tripDirection:       map[string]string{},
		tripShapeID:         map[string]string{},
		tripBlockID:         map[string]string{},
		TripStopSeq:         map[string][]string{},
		tripStopIdx:         map[string]map[string]int{},
		stopNames:           map[string]string{},
		stopCoord:           map[string][2]float64{},
		ShapePoints:         map[string][][2]float64{},
		ShapeCumKM:          map[string][]float64{},
		stopTimePickupType:  map[string]map[string]int{},
		stopTimeDropOffType: map[string]map[string]int{},
		stopTimeArrival:     map[string]map[string]string{},
		stopTimeDeparture:   map[string]map[string]string{},
	}, nil
}

// NewGTFSIndexFromConfig creates and loads a GTFS index from configuration
func NewGTFSIndexFromConfig(cfg config.GTFSConfig) (*GTFSIndex, error) {
	g, _ := NewGTFSIndex("", "")
	g.agencyID = cfg.AgencyID
	if cfg.StaticURL != "" {
		if err := g.loadFromStaticZip(cfg.StaticURL); err != nil {
			return g, err
		}
		return g, nil
	}
	return g, nil
}

// Accessor methods
func (g *GTFSIndex) GetAgencyTimezone(agencyID string) string {
	if g.agencyTZ != "" {
		return g.agencyTZ
	}
	return "America/New_York"
}

func (g *GTFSIndex) GetOriginStopIDForTrip(gtfsTripKey string) string {
	return g.tripOriginStop[gtfsTripKey]
}

func (g *GTFSIndex) GetDestinationStopIDForTrip(gtfsTripKey string) string {
	return g.tripDestStop[gtfsTripKey]
}

func (g *GTFSIndex) GetTripHeadsign(gtfsTripKey string) string { return g.tripHeadsign[gtfsTripKey] }

func (g *GTFSIndex) GetShapeIDForTrip(gtfsTripKey string) string { return g.tripShapeID[gtfsTripKey] }

func (g *GTFSIndex) GetFullTripIDForTrip(gtfsTripKey string) string { return gtfsTripKey }

func (g *GTFSIndex) GetBlockIDForTrip(gtfsTripKey string) string { return g.tripBlockID[gtfsTripKey] }

func (g *GTFSIndex) GetRouteShortName(routeID string) string { return g.routeShortNames[routeID] }

func (g *GTFSIndex) GetRouteType(routeID string) int { return g.routeTypes[routeID] }

func (g *GTFSIndex) GetAgencyName() string { return g.agencyName }

func (g *GTFSIndex) GetStopName(stopID string) string { return g.stopNames[stopID] }

func (g *GTFSIndex) GetPreviousStopIDOfStopForTrip(gtfsTripKey, stopID string) string {
	if m, ok := g.tripStopIdx[gtfsTripKey]; ok {
		if idx, ok2 := m[stopID]; ok2 {
			if idx > 0 {
				return g.TripStopSeq[gtfsTripKey][idx-1]
			}
			return ""
		}
	}
	return ""
}

// GetPickupType returns the pickup_type for a stop in a trip (0=regular, 1=none, 2=phone, 3=coordinate)
func (g *GTFSIndex) GetPickupType(gtfsTripKey, stopID string) int {
	if m, ok := g.stopTimePickupType[gtfsTripKey]; ok {
		return m[stopID]
	}
	return 0 // Default: regular pickup
}

// GetDropOffType returns the drop_off_type for a stop in a trip (0=regular, 1=none, 2=phone, 3=coordinate)
func (g *GTFSIndex) GetDropOffType(gtfsTripKey, stopID string) int {
	if m, ok := g.stopTimeDropOffType[gtfsTripKey]; ok {
		return m[stopID]
	}
	return 0 // Default: regular drop off
}

// GetArrivalTime returns the static arrival_time string (HH:MM:SS) for a stop in a trip
func (g *GTFSIndex) GetArrivalTime(gtfsTripKey, stopID string) string {
	if m, ok := g.stopTimeArrival[gtfsTripKey]; ok {
		return m[stopID]
	}
	return ""
}

// GetDepartureTime returns the static departure_time string (HH:MM:SS) for a stop in a trip
func (g *GTFSIndex) GetDepartureTime(gtfsTripKey, stopID string) string {
	if m, ok := g.stopTimeDeparture[gtfsTripKey]; ok {
		return m[stopID]
	}
	return ""
}

func (g *GTFSIndex) GetSliceShapeForTrip(gtfsTripKey string, startSeg, endSeg int) []Waypoint {
	shapeID := g.GetShapeIDForTrip(gtfsTripKey)
	pts := g.ShapePoints[shapeID]
	if len(pts) == 0 {
		return nil
	}
	if startSeg > endSeg {
		startSeg, endSeg = endSeg, startSeg
	}
	if startSeg < 0 {
		startSeg = 0
	}
	if endSeg >= len(pts) {
		endSeg = len(pts) - 1
	}
	if startSeg >= len(pts) || startSeg > endSeg {
		return nil
	}
	out := make([]Waypoint, 0, endSeg-startSeg+1)
	for i := startSeg; i <= endSeg; i++ {
		out = append(out, Waypoint{Longitude: pts[i][0], Latitude: pts[i][1]})
	}
	return out
}

func (g *GTFSIndex) GetSnappedCoordinatesOfStopForTrip(gtfsTripKey, stopID string) []float64 {
	shapeID := g.GetShapeIDForTrip(gtfsTripKey)
	pts := g.ShapePoints[shapeID]
	coord, ok := g.stopCoord[stopID]
	if !ok || len(pts) < 2 {
		if ok {
			return []float64{coord[0], coord[1]}
		}
		return nil
	}
	_, _, snapped := NearestSegmentProjection(pts, coord)
	return []float64{snapped[0], snapped[1]}
}

func (g *GTFSIndex) GetShapeSegmentNumberOfStopForTrip(gtfsTripKey, stopID string) int {
	shapeID := g.GetShapeIDForTrip(gtfsTripKey)
	pts := g.ShapePoints[shapeID]
	coord, ok := g.stopCoord[stopID]
	if !ok || len(pts) < 2 {
		return -1
	}
	idx, _, _ := NearestSegmentProjection(pts, coord)
	return idx
}

func (g *GTFSIndex) TripIsAScheduledTrip(gtfsTripKey string) bool {
	_, ok := g.tripToRoute[gtfsTripKey]
	return ok
}

func (g *GTFSIndex) TripsHasSpatialData(gtfsTripKey string) bool {
	sh := g.GetShapeIDForTrip(gtfsTripKey)
	return sh != "" && len(g.ShapePoints[sh]) > 1
}

func (g *GTFSIndex) GetAllStops() []string {
	keys := make([]string, 0, len(g.stopNames))
	for k := range g.stopNames {
		keys = append(keys, k)
	}
	return keys
}

func (g *GTFSIndex) GetAllRoutes() []string {
	keys := make([]string, 0, len(g.routeShortNames))
	for k := range g.routeShortNames {
		keys = append(keys, k)
	}
	return keys
}

func (g *GTFSIndex) GetAllAgencyIDs() []string {
	if g.agencyID != "" {
		return []string{g.agencyID}
	}
	return []string{}
}

func (g *GTFSIndex) GetRouteIDForTrip(gtfsTripKey string) string { return g.tripToRoute[gtfsTripKey] }

func (g *GTFSIndex) GetDirectionIDForTrip(gtfsTripKey string) string {
	return g.tripDirection[gtfsTripKey]
}
