package gtfsrtsiri

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Waypoint struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

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
	tripStopSeq     map[string][]string       // trip_id -> ordered stop_ids
	tripStopIdx     map[string]map[string]int // trip_id -> stop_id -> index
	stopNames       map[string]string         // stop_id -> name
	stopCoord       map[string][2]float64     // stop_id -> [lon,lat]
	shapePoints     map[string][][2]float64   // shape_id -> ordered points [lon,lat]
	shapeCumKM      map[string][]float64      // shape_id -> cumulative km at each point
	// New fields for ET support
	stopTimePickupType  map[string]map[string]int    // trip_id -> stop_id -> pickup_type
	stopTimeDropOffType map[string]map[string]int    // trip_id -> stop_id -> drop_off_type
	stopTimeArrival     map[string]map[string]string // trip_id -> stop_id -> arrival_time (HH:MM:SS)
	stopTimeDeparture   map[string]map[string]string // trip_id -> stop_id -> departure_time (HH:MM:SS)
}

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
		tripStopSeq:         map[string][]string{},
		tripStopIdx:         map[string]map[string]int{},
		stopNames:           map[string]string{},
		stopCoord:           map[string][2]float64{},
		shapePoints:         map[string][][2]float64{},
		shapeCumKM:          map[string][]float64{},
		stopTimePickupType:  map[string]map[string]int{},
		stopTimeDropOffType: map[string]map[string]int{},
		stopTimeArrival:     map[string]map[string]string{},
		stopTimeDeparture:   map[string]map[string]string{},
	}, nil
}

func NewGTFSIndexFromConfig(cfg GTFSConfig) (*GTFSIndex, error) {
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

func (g *GTFSIndex) loadFromStaticZip(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	tmp, err := os.CreateTemp("", "gtfs-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	zr, err := zip.OpenReader(tmp.Name())
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if name == "routes.txt" || name == "trips.txt" || name == "stops.txt" || name == "stop_times.txt" || name == "agency.txt" || name == "shapes.txt" {
			if err := g.consumeCSV(f); err != nil {
				return err
			}
		}
	}
	// After shapes ingested, compute cumulative distances
	for shapeID, pts := range g.shapePoints {
		g.shapeCumKM[shapeID] = cumulativeKM(pts)
	}
	return nil
}

// loadFromLocalZip opens a local GTFS zip file and consumes required CSVs.
func (g *GTFSIndex) loadFromLocalZip(path string) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if name == "routes.txt" || name == "trips.txt" || name == "stops.txt" || name == "stop_times.txt" || name == "agency.txt" || name == "shapes.txt" {
			if err := g.consumeCSV(f); err != nil {
				return err
			}
		}
	}
	for shapeID, pts := range g.shapePoints {
		g.shapeCumKM[shapeID] = cumulativeKM(pts)
	}
	return nil
}

// Utility converters for flexible JSON values
func toStringFallback(v any, fallback string) string {
	switch t := v.(type) {
	case string:
		if t != "" {
			return t
		}
	case float64:
		return strconv.Itoa(int(t))
	case json.Number:
		if i, err := strconv.Atoi(t.String()); err == nil {
			return strconv.Itoa(i)
		}
	}
	return fallback
}

func toFloat(v any) (float64, error) {
	switch t := v.(type) {
	case float64:
		return t, nil
	case string:
		return strconv.ParseFloat(t, 64)
	case json.Number:
		return t.Float64()
	default:
		return 0, errors.New("not a float")
	}
}

func toInt(v any) (int, error) {
	switch t := v.(type) {
	case float64:
		return int(t), nil
	case string:
		return strconv.Atoi(t)
	case json.Number:
		i64, err := t.Int64()
		return int(i64), err
	default:
		return 0, errors.New("not an int")
	}
}

func (g *GTFSIndex) consumeCSV(f *zip.File) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	csvr := csv.NewReader(r)
	rec, err := csvr.ReadAll()
	if err != nil {
		return err
	}
	if len(rec) == 0 {
		return nil
	}
	head := rec[0]
	idx := func(col string) int {
		for i, h := range head {
			if strings.EqualFold(h, col) {
				return i
			}
		}
		return -1
	}
	switch strings.ToLower(f.Name) {
	case "routes.txt":
		rID := idx("route_id")
		rSN := idx("route_short_name")
		rType := idx("route_type")
		for _, row := range rec[1:] {
			if rID >= 0 && rSN >= 0 {
				g.routeShortNames[row[rID]] = row[rSN]
			}
			if rID >= 0 && rType >= 0 {
				if typeInt, err := strconv.Atoi(row[rType]); err == nil {
					g.routeTypes[row[rID]] = typeInt
				}
			}
		}
	case "trips.txt":
		rID := idx("route_id")
		tID := idx("trip_id")
		hs := idx("trip_headsign")
		dir := idx("direction_id")
		sh := idx("shape_id")
		blk := idx("block_id")
		for _, row := range rec[1:] {
			if tID >= 0 && rID >= 0 {
				g.tripToRoute[row[tID]] = row[rID]
			}
			if tID >= 0 && hs >= 0 {
				g.tripHeadsign[row[tID]] = row[hs]
			}
			if tID >= 0 && dir >= 0 {
				g.tripDirection[row[tID]] = row[dir]
			}
			if tID >= 0 && sh >= 0 {
				g.tripShapeID[row[tID]] = row[sh]
			}
			if tID >= 0 && blk >= 0 {
				g.tripBlockID[row[tID]] = row[blk]
			}
		}
	case "stops.txt":
		sID := idx("stop_id")
		sN := idx("stop_name")
		sLat := idx("stop_lat")
		sLon := idx("stop_lon")
		for _, row := range rec[1:] {
			if sID >= 0 && sN >= 0 {
				g.stopNames[row[sID]] = row[sN]
			}
			if sID >= 0 && sLat >= 0 && sLon >= 0 {
				lat, _ := strconv.ParseFloat(row[sLat], 64)
				lon, _ := strconv.ParseFloat(row[sLon], 64)
				g.stopCoord[row[sID]] = [2]float64{lon, lat}
			}
		}
	case "stop_times.txt":
		tID := idx("trip_id")
		sID := idx("stop_id")
		sq := idx("stop_sequence")
		arrTime := idx("arrival_time")
		depTime := idx("departure_time")
		pickupType := idx("pickup_type")
		dropOffType := idx("drop_off_type")
		if tID < 0 || sID < 0 || sq < 0 {
			return nil
		}
		tmp := map[string][]struct {
			stop        string
			seq         int
			arrTime     string
			depTime     string
			pickupType  int
			dropOffType int
		}{}
		for _, row := range rec[1:] {
			trip := row[tID]
			stop := row[sID]
			seq, _ := strconv.Atoi(row[sq])
			arrT := ""
			if arrTime >= 0 && arrTime < len(row) {
				arrT = row[arrTime]
			}
			depT := ""
			if depTime >= 0 && depTime < len(row) {
				depT = row[depTime]
			}
			pickup := 0
			if pickupType >= 0 && pickupType < len(row) && row[pickupType] != "" {
				pickup, _ = strconv.Atoi(row[pickupType])
			}
			dropOff := 0
			if dropOffType >= 0 && dropOffType < len(row) && row[dropOffType] != "" {
				dropOff, _ = strconv.Atoi(row[dropOffType])
			}
			tmp[trip] = append(tmp[trip], struct {
				stop        string
				seq         int
				arrTime     string
				depTime     string
				pickupType  int
				dropOffType int
			}{stop, seq, arrT, depT, pickup, dropOff})
		}
		for trip, arr := range tmp {
			sort.Slice(arr, func(i, j int) bool { return arr[i].seq < arr[j].seq })
			// first/last
			if len(arr) > 0 {
				g.tripOriginStop[trip] = arr[0].stop
				g.tripDestStop[trip] = arr[len(arr)-1].stop
			}
			// Initialize maps for this trip
			g.stopTimePickupType[trip] = make(map[string]int)
			g.stopTimeDropOffType[trip] = make(map[string]int)
			g.stopTimeArrival[trip] = make(map[string]string)
			g.stopTimeDeparture[trip] = make(map[string]string)
			// stop sequence + index map + stop times data
			seqStops := make([]string, 0, len(arr))
			idxMap := make(map[string]int, len(arr))
			for i, v := range arr {
				seqStops = append(seqStops, v.stop)
				if _, ok := idxMap[v.stop]; !ok {
					idxMap[v.stop] = i
				}
				// Store pickup/dropoff types and times
				g.stopTimePickupType[trip][v.stop] = v.pickupType
				g.stopTimeDropOffType[trip][v.stop] = v.dropOffType
				if v.arrTime != "" {
					g.stopTimeArrival[trip][v.stop] = v.arrTime
				}
				if v.depTime != "" {
					g.stopTimeDeparture[trip][v.stop] = v.depTime
				}
			}
			g.tripStopSeq[trip] = seqStops
			g.tripStopIdx[trip] = idxMap
		}
	case "agency.txt":
		agID := idx("agency_id")
		agTZ := idx("agency_timezone")
		agName := idx("agency_name")
		if len(rec) > 1 {
			if agID >= 0 && g.agencyID == "" {
				g.agencyID = rec[1][agID]
			}
			if agTZ >= 0 {
				g.agencyTZ = rec[1][agTZ]
			}
			if agName >= 0 {
				g.agencyName = rec[1][agName]
			}
		}
	case "shapes.txt":
		sh := idx("shape_id")
		latIdx := idx("shape_pt_lat")
		lonIdx := idx("shape_pt_lon")
		seqIdx := idx("shape_pt_sequence")
		if sh < 0 || latIdx < 0 || lonIdx < 0 || seqIdx < 0 {
			return nil
		}
		tmp := map[string][]struct {
			lon, lat float64
			seq      int
		}{}
		for _, row := range rec[1:] {
			shapeID := row[sh]
			lat, _ := strconv.ParseFloat(row[latIdx], 64)
			lon, _ := strconv.ParseFloat(row[lonIdx], 64)
			seq, _ := strconv.Atoi(row[seqIdx])
			tmp[shapeID] = append(tmp[shapeID], struct {
				lon, lat float64
				seq      int
			}{lon, lat, seq})
		}
		for shapeID, arr := range tmp {
			sort.Slice(arr, func(i, j int) bool { return arr[i].seq < arr[j].seq })
			pts := make([][2]float64, len(arr))
			for i, p := range arr {
				pts[i] = [2]float64{p.lon, p.lat}
			}
			g.shapePoints[shapeID] = pts
		}
	}
	return nil
}

// Removed: index export helpers and indexer path (library is ZIP-only, in-memory)

// Optionally used for debugging loaded data
func (g *GTFSIndex) dumpDebugJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

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
func (g *GTFSIndex) GetTripHeadsign(gtfsTripKey string) string      { return g.tripHeadsign[gtfsTripKey] }
func (g *GTFSIndex) GetShapeIDForTrip(gtfsTripKey string) string    { return g.tripShapeID[gtfsTripKey] }
func (g *GTFSIndex) GetFullTripIDForTrip(gtfsTripKey string) string { return gtfsTripKey }
func (g *GTFSIndex) GetBlockIDForTrip(gtfsTripKey string) string    { return g.tripBlockID[gtfsTripKey] }
func (g *GTFSIndex) GetRouteShortName(routeID string) string        { return g.routeShortNames[routeID] }
func (g *GTFSIndex) GetRouteType(routeID string) int                { return g.routeTypes[routeID] }
func (g *GTFSIndex) GetAgencyName() string                          { return g.agencyName }
func (g *GTFSIndex) GetStopName(stopID string) string               { return g.stopNames[stopID] }

func (g *GTFSIndex) GetStopDistanceAlongRouteForTripInMeters(gtfsTripKey, stopID string) float64 {
	km := g.GetStopDistanceAlongRouteForTripInKilometers(gtfsTripKey, stopID)
	if math.IsNaN(km) {
		return 0
	}
	return km * 1000
}

func (g *GTFSIndex) GetStopDistanceAlongRouteForTripInKilometers(gtfsTripKey, stopID string) float64 {
	shapeID := g.GetShapeIDForTrip(gtfsTripKey)
	if shapeID == "" {
		return 0
	}
	pts := g.shapePoints[shapeID]
	if len(pts) < 2 {
		return 0
	}
	coord, ok := g.stopCoord[stopID]
	if !ok {
		return 0
	}
	segIdx, t, _ := nearestSegmentProjection(pts, coord)
	cum := g.shapeCumKM[shapeID]
	if segIdx < 0 || segIdx >= len(cum) {
		return 0
	}
	if segIdx == len(pts)-1 {
		return cum[segIdx]
	}
	// add fractional distance within the segment
	segKM := haversineKM(pts[segIdx][1], pts[segIdx][0], pts[segIdx+1][1], pts[segIdx+1][0])
	return cum[segIdx] + t*segKM
}

func (g *GTFSIndex) GetPreviousStopIDOfStopForTrip(gtfsTripKey, stopID string) string {
	if m, ok := g.tripStopIdx[gtfsTripKey]; ok {
		if idx, ok2 := m[stopID]; ok2 {
			if idx > 0 {
				return g.tripStopSeq[gtfsTripKey][idx-1]
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
	pts := g.shapePoints[shapeID]
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
	pts := g.shapePoints[shapeID]
	coord, ok := g.stopCoord[stopID]
	if !ok || len(pts) < 2 {
		if ok {
			return []float64{coord[0], coord[1]}
		}
		return nil
	}
	_, _, snapped := nearestSegmentProjection(pts, coord)
	return []float64{snapped[0], snapped[1]}
}
func (g *GTFSIndex) GetShapeSegmentNumberOfStopForTrip(gtfsTripKey, stopID string) int {
	shapeID := g.GetShapeIDForTrip(gtfsTripKey)
	pts := g.shapePoints[shapeID]
	coord, ok := g.stopCoord[stopID]
	if !ok || len(pts) < 2 {
		return -1
	}
	idx, _, _ := nearestSegmentProjection(pts, coord)
	return idx
}
func (g *GTFSIndex) TripIsAScheduledTrip(gtfsTripKey string) bool {
	_, ok := g.tripToRoute[gtfsTripKey]
	return ok
}
func (g *GTFSIndex) TripsHasSpatialData(gtfsTripKey string) bool {
	sh := g.GetShapeIDForTrip(gtfsTripKey)
	return sh != "" && len(g.shapePoints[sh]) > 1
}

// GetCoordinateAtDistanceForTrip returns a lon,lat point on the trip's shape at a target distance in KM
func (g *GTFSIndex) GetCoordinateAtDistanceForTrip(gtfsTripKey string, targetKM float64) (float64, float64, bool) {
	shapeID := g.GetShapeIDForTrip(gtfsTripKey)
	pts := g.shapePoints[shapeID]
	cum := g.shapeCumKM[shapeID]
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

// Helpers
func cumulativeKM(pts [][2]float64) []float64 {
	cum := make([]float64, len(pts))
	if len(pts) == 0 {
		return cum
	}
	cum[0] = 0
	for i := 1; i < len(pts); i++ {
		cum[i] = cum[i-1] + haversineKM(pts[i-1][1], pts[i-1][0], pts[i][1], pts[i][0])
	}
	return cum
}

func nearestPointIndex(pts [][2]float64, coord [2]float64) int {
	best := -1
	bestD := math.MaxFloat64
	for i, p := range pts {
		d := haversineKM(coord[1], coord[0], p[1], p[0])
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
func nearestSegmentProjection(pts [][2]float64, coord [2]float64) (int, float64, [2]float64) {
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

func haversineKM(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	la1 := lat1 * math.Pi / 180
	la2 := lat2 * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(la1)*math.Cos(la2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
