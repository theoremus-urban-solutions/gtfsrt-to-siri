package gtfs

// GTFSIndex stores GTFS static data in memory for fast lookups.
// This index is data-source agnostic - it accepts raw zip data
// and does NOT handle HTTP downloads or file paths.
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
	tripBlockID     map[string]string         // trip_id -> block_id
	TripStopSeq map[string][]string       // trip_id -> ordered stop_ids
	tripStopIdx map[string]map[string]int // trip_id -> stop_id -> index
	stopNames   map[string]string         // stop_id -> name
	StopCoord   map[string][2]float64     // stop_id -> [lon,lat]
	// Fields for ET support
	stopTimes map[string]map[string]StopTime // trip_id -> stop_id -> StopTime
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

func (g *GTFSIndex) GetFullTripIDForTrip(gtfsTripKey string) string { return gtfsTripKey }

func (g *GTFSIndex) GetBlockIDForTrip(gtfsTripKey string) string { return g.tripBlockID[gtfsTripKey] }

func (g *GTFSIndex) GetRouteShortName(routeID string) string { return g.routeShortNames[routeID] }

func (g *GTFSIndex) GetRouteType(routeID string) int { return g.routeTypes[routeID] }

func (g *GTFSIndex) GetRouteTypeWithExists(routeID string) (int, bool) {
	routeType, exists := g.routeTypes[routeID]
	return routeType, exists
}

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
	if m, ok := g.stopTimes[gtfsTripKey]; ok {
		if st, ok2 := m[stopID]; ok2 {
			return int(st.PickupType)
		}
	}
	return 0 // Default: regular pickup
}

// GetDropOffType returns the drop_off_type for a stop in a trip (0=regular, 1=none, 2=phone, 3=coordinate)
func (g *GTFSIndex) GetDropOffType(gtfsTripKey, stopID string) int {
	if m, ok := g.stopTimes[gtfsTripKey]; ok {
		if st, ok2 := m[stopID]; ok2 {
			return int(st.DropOffType)
		}
	}
	return 0 // Default: regular drop off
}

// GetArrivalTime returns the static arrival_time string (HH:MM:SS) for a stop in a trip
func (g *GTFSIndex) GetArrivalTime(gtfsTripKey, stopID string) string {
	if m, ok := g.stopTimes[gtfsTripKey]; ok {
		if st, ok2 := m[stopID]; ok2 {
			return st.ArrivalTime
		}
	}
	return ""
}

// GetDepartureTime returns the static departure_time string (HH:MM:SS) for a stop in a trip
func (g *GTFSIndex) GetDepartureTime(gtfsTripKey, stopID string) string {
	if m, ok := g.stopTimes[gtfsTripKey]; ok {
		if st, ok2 := m[stopID]; ok2 {
			return st.DepartureTime
		}
	}
	return ""
}

func (g *GTFSIndex) TripIsAScheduledTrip(gtfsTripKey string) bool {
	_, ok := g.tripToRoute[gtfsTripKey]
	return ok
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
