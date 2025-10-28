package gtfs

// GTFSIndex stores GTFS static data in memory for fast lookups.
// This index is data-source agnostic - it accepts raw zip data
// and does NOT handle HTTP downloads or file paths.
//
// All fields are exported to support serialization via encoding/gob for caching.
// Use SerializeIndex/DeserializeIndex for disk-based caching.
//
// Thread safety: GTFSIndex is safe for concurrent read access after construction.
// Multiple goroutines can safely read from a shared GTFSIndex instance.
type GTFSIndex struct {
	AgencyID        string                             // Agency ID from config or agency.txt
	AgencyTZ        string                             // Agency timezone from agency.txt
	AgencyName      string                             // Agency name from agency.txt
	RouteShortNames map[string]string                  // route_id -> short_name
	RouteTypes      map[string]int                     // route_id -> route_type (GTFS enum)
	Routes          map[string]struct{}                // route existence set
	TripToRoute     map[string]string                  // trip_id -> route_id
	TripHeadsign    map[string]string                  // trip_id -> headsign
	TripOriginStop  map[string]string                  // trip_id -> first stop_id
	TripDestStop    map[string]string                  // trip_id -> last stop_id
	TripDirection   map[string]string                  // trip_id -> direction_id ("0"|"1")
	TripBlockID     map[string]string                  // trip_id -> block_id
	TripStopSeq     map[string][]string                // trip_id -> ordered stop_ids (exported for caching)
	TripStopIdx     map[string]map[string]int          // trip_id -> stop_id -> index
	StopNames       map[string]string                  // stop_id -> name
	StopCoord       map[string][2]float64              // stop_id -> [lon,lat] (exported for caching)
	StopTimes       map[string]map[string]StopTime     // trip_id -> stop_id -> StopTime (for ET support)
}

// Accessor methods
func (g *GTFSIndex) GetAgencyTimezone(agencyID string) string {
	if g.AgencyTZ != "" {
		return g.AgencyTZ
	}
	return "America/New_York"
}

func (g *GTFSIndex) GetOriginStopIDForTrip(gtfsTripKey string) string {
	return g.TripOriginStop[gtfsTripKey]
}

func (g *GTFSIndex) GetDestinationStopIDForTrip(gtfsTripKey string) string {
	return g.TripDestStop[gtfsTripKey]
}

func (g *GTFSIndex) GetTripHeadsign(gtfsTripKey string) string { return g.TripHeadsign[gtfsTripKey] }

func (g *GTFSIndex) GetFullTripIDForTrip(gtfsTripKey string) string { return gtfsTripKey }

func (g *GTFSIndex) GetBlockIDForTrip(gtfsTripKey string) string { return g.TripBlockID[gtfsTripKey] }

func (g *GTFSIndex) GetRouteShortName(routeID string) string { return g.RouteShortNames[routeID] }

func (g *GTFSIndex) GetRouteType(routeID string) int { return g.RouteTypes[routeID] }

func (g *GTFSIndex) GetRouteTypeWithExists(routeID string) (int, bool) {
	routeType, exists := g.RouteTypes[routeID]
	return routeType, exists
}

func (g *GTFSIndex) GetAgencyName() string { return g.AgencyName }

func (g *GTFSIndex) GetStopName(stopID string) string { return g.StopNames[stopID] }

func (g *GTFSIndex) GetPreviousStopIDOfStopForTrip(gtfsTripKey, stopID string) string {
	if m, ok := g.TripStopIdx[gtfsTripKey]; ok {
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
	if m, ok := g.StopTimes[gtfsTripKey]; ok {
		if st, ok2 := m[stopID]; ok2 {
			return int(st.PickupType)
		}
	}
	return 0 // Default: regular pickup
}

// GetDropOffType returns the drop_off_type for a stop in a trip (0=regular, 1=none, 2=phone, 3=coordinate)
func (g *GTFSIndex) GetDropOffType(gtfsTripKey, stopID string) int {
	if m, ok := g.StopTimes[gtfsTripKey]; ok {
		if st, ok2 := m[stopID]; ok2 {
			return int(st.DropOffType)
		}
	}
	return 0 // Default: regular drop off
}

// GetArrivalTime returns the static arrival_time string (HH:MM:SS) for a stop in a trip
func (g *GTFSIndex) GetArrivalTime(gtfsTripKey, stopID string) string {
	if m, ok := g.StopTimes[gtfsTripKey]; ok {
		if st, ok2 := m[stopID]; ok2 {
			return st.ArrivalTime
		}
	}
	return ""
}

// GetDepartureTime returns the static departure_time string (HH:MM:SS) for a stop in a trip
func (g *GTFSIndex) GetDepartureTime(gtfsTripKey, stopID string) string {
	if m, ok := g.StopTimes[gtfsTripKey]; ok {
		if st, ok2 := m[stopID]; ok2 {
			return st.DepartureTime
		}
	}
	return ""
}

func (g *GTFSIndex) TripIsAScheduledTrip(gtfsTripKey string) bool {
	_, ok := g.TripToRoute[gtfsTripKey]
	return ok
}

func (g *GTFSIndex) GetAllStops() []string {
	keys := make([]string, 0, len(g.StopNames))
	for k := range g.StopNames {
		keys = append(keys, k)
	}
	return keys
}

func (g *GTFSIndex) GetAllRoutes() []string {
	keys := make([]string, 0, len(g.RouteShortNames))
	for k := range g.RouteShortNames {
		keys = append(keys, k)
	}
	return keys
}

func (g *GTFSIndex) GetAllAgencyIDs() []string {
	if g.AgencyID != "" {
		return []string{g.AgencyID}
	}
	return []string{}
}

func (g *GTFSIndex) GetRouteIDForTrip(gtfsTripKey string) string { return g.TripToRoute[gtfsTripKey] }

func (g *GTFSIndex) GetDirectionIDForTrip(gtfsTripKey string) string {
	return g.TripDirection[gtfsTripKey]
}
