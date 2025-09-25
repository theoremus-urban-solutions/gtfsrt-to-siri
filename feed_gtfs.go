package gtfsrtsiri

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
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
	routeShortNames map[string]string         // route_id -> short_name
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
}

func NewGTFSIndex(indexedSchedulePath, indexedSpatialPath string) (*GTFSIndex, error) {
	return &GTFSIndex{
		routeShortNames: map[string]string{},
		routes:          map[string]struct{}{},
		tripToRoute:     map[string]string{},
		tripHeadsign:    map[string]string{},
		tripOriginStop:  map[string]string{},
		tripDestStop:    map[string]string{},
		tripDirection:   map[string]string{},
		tripShapeID:     map[string]string{},
		tripBlockID:     map[string]string{},
		tripStopSeq:     map[string][]string{},
		tripStopIdx:     map[string]map[string]int{},
		stopNames:       map[string]string{},
		stopCoord:       map[string][2]float64{},
		shapePoints:     map[string][][2]float64{},
		shapeCumKM:      map[string][]float64{},
	}, nil
}

func NewGTFSIndexFromConfig(cfg GTFSConfig) (*GTFSIndex, error) {
	g, _ := NewGTFSIndex(cfg.IndexedSchedulePath, cfg.IndexedSpatialPath)
	g.agencyID = cfg.AgencyID
	if cfg.IndexedSchedulePath != "" && cfg.IndexedSpatialPath != "" {
		if err := g.loadFromIndexedJSON(cfg.IndexedSchedulePath, cfg.IndexedSpatialPath); err != nil {
			return g, err
		}
		return g, nil
	}
	if cfg.StaticURL != "" {
		if err := g.loadFromStaticZip(cfg.StaticURL); err != nil {
			return g, err
		}
		return g, nil
	}
	if cfg.IndexedSchedulePath != "" {
		if err := g.loadFromIndexedSchedule(cfg.IndexedSchedulePath); err != nil {
			return g, err
		}
	}
	if cfg.IndexedSpatialPath != "" {
		if err := g.loadFromIndexedSpatial(cfg.IndexedSpatialPath); err != nil {
			return g, err
		}
	}
	return g, nil
}

func (g *GTFSIndex) loadFromIndexedJSON(schedulePath, spatialPath string) error {
	if err := g.loadFromIndexedSchedule(schedulePath); err != nil {
		return err
	}
	if err := g.loadFromIndexedSpatial(spatialPath); err != nil {
		return err
	}
	return nil
}

func (g *GTFSIndex) loadFromIndexedSchedule(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var container map[string]json.RawMessage
	if err := json.Unmarshal(raw, &container); err != nil {
		return err
	}
	if rawAgency, ok := container["agency"]; ok {
		var agencies map[string]json.RawMessage
		if err := json.Unmarshal(rawAgency, &agencies); err == nil {
			for _, entry := range agencies {
				var ag struct {
					AgencyID       string `json:"agency_id"`
					AgencyTimezone string `json:"agency_timezone"`
				}
				if err := json.Unmarshal(entry, &ag); err != nil {
					continue
				}
				if g.agencyID == "" && ag.AgencyID != "" {
					g.agencyID = ag.AgencyID
				}
				if ag.AgencyTimezone != "" {
					g.agencyTZ = ag.AgencyTimezone
				}
			}
		}
	}
	if rawRoutes, ok := container["routes"]; ok {
		var routes map[string]json.RawMessage
		if err := json.Unmarshal(rawRoutes, &routes); err == nil {
			for key, entry := range routes {
				var r map[string]any
				if err := json.Unmarshal(entry, &r); err != nil {
					return err
				}
				rid := toStringFallback(r["route_id"], key)
				if rid == "" {
					rid = key
				}
				g.routes[rid] = struct{}{}
				if sn := toStringFallback(r["route_short_name"], ""); sn != "" {
					g.routeShortNames[rid] = sn
				}
			}
		}
	}
	if rawStops, ok := container["stops"]; ok {
		var stops map[string]json.RawMessage
		if err := json.Unmarshal(rawStops, &stops); err == nil {
			for key, entry := range stops {
				var s map[string]any
				if err := json.Unmarshal(entry, &s); err != nil {
					return err
				}
				sid := toStringFallback(s["stop_id"], key)
				if sid == "" {
					sid = key
				}
				g.stopNames[sid] = toStringFallback(s["stop_name"], "")
				lat, _ := toFloat(s["stop_lat"])
				lon, _ := toFloat(s["stop_lon"])
				g.stopCoord[sid] = [2]float64{lon, lat}
			}
		}
	}
	if rawTrips, ok := container["trips"]; ok {
		var trips map[string]json.RawMessage
		if err := json.Unmarshal(rawTrips, &trips); err == nil {
			for key, entry := range trips {
				var t map[string]any
				if err := json.Unmarshal(entry, &t); err != nil {
					return err
				}
				g.tripToRoute[key] = toStringFallback(t["route_id"], "")
				g.tripHeadsign[key] = toStringFallback(t["trip_headsign"], "")
				if dirVal, ok := t["direction_id"]; ok {
					if di, err := toInt(dirVal); err == nil {
						g.tripDirection[key] = strconv.Itoa(di)
					}
				}
				g.tripShapeID[key] = toStringFallback(t["shape_id"], "")
				g.tripBlockID[key] = toStringFallback(t["block_id"], "")
			}
		}
	}
	if rawStopTimes, ok := container["stopTimes"]; ok {
		var stopTimes map[string]json.RawMessage
		if err := json.Unmarshal(rawStopTimes, &stopTimes); err == nil {
			for tripKey, entry := range stopTimes {
				var seq []map[string]any
				if err := json.Unmarshal(entry, &seq); err != nil {
					return err
				}
				sort.Slice(seq, func(i, j int) bool {
					si, _ := toInt(seq[i]["stop_sequence"])
					sj, _ := toInt(seq[j]["stop_sequence"])
					return si < sj
				})
				stops := make([]string, 0, len(seq))
				idxs := make(map[string]int, len(seq))
				for idx, st := range seq {
					sid := toStringFallback(st["stop_id"], "")
					stops = append(stops, sid)
					if _, ok := idxs[sid]; !ok {
						idxs[sid] = idx
					}
				}
				if len(stops) > 0 {
					g.tripOriginStop[tripKey] = stops[0]
					g.tripDestStop[tripKey] = stops[len(stops)-1]
				}
				g.tripStopSeq[tripKey] = stops
				g.tripStopIdx[tripKey] = idxs
			}
		}
	}
	return nil
}

func (g *GTFSIndex) loadFromIndexedSpatial(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var payload struct {
		Paths map[string][]struct {
			Latitude     float64 `json:"latitude"`
			Longitude    float64 `json:"longitude"`
			DistTraveled float64 `json:"dist_traveled"`
		} `json:"paths"`
		StopProjectionsTable [][]struct {
			StopID        string    `json:"stop_id"`
			SegmentNum    int       `json:"segmentNum"`
			SnappedCoords []float64 `json:"snapped_coords"`
			SnappedDist   float64   `json:"snapped_dist_traveled"`
		} `json:"stopProjectionsTable"`
		TripKeyToProjectionTableIndex map[string]int `json:"tripKeyToProjectionsTableIndex"`
	}
	if err := json.Unmarshal(b, &payload); err != nil {
		return err
	}
	g.shapePoints = map[string][][2]float64{}
	g.shapeCumKM = map[string][]float64{}
	for shapeID, pts := range payload.Paths {
		coords := make([][2]float64, len(pts))
		cum := make([]float64, len(pts))
		prev := float64(0)
		for i, p := range pts {
			coords[i] = [2]float64{p.Longitude, p.Latitude}
			if i == 0 {
				cum[i] = 0
			} else {
				prev += haversineKM(coords[i-1][1], coords[i-1][0], coords[i][1], coords[i][0])
				cum[i] = prev
			}
		}
		g.shapePoints[shapeID] = coords
		g.shapeCumKM[shapeID] = cum
	}
	// stop projections used for snapped coords if required
	// Currently we rely on nearestSegmentProjection; we can later integrate stop projections
	return nil
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

// ExportIndexedSchedule emits a Node-like indexedScheduleData JSON blob.
func (g *GTFSIndex) ExportIndexedSchedule() ([]byte, error) {
	type agencyRec struct {
		AgencyID       string `json:"agency_id"`
		AgencyTimezone string `json:"agency_timezone"`
	}
	sched := map[string]any{}
	// agency map keyed by agency_id for parity shape
	agMap := map[string]agencyRec{}
	if g.agencyID != "" {
		agMap[g.agencyID] = agencyRec{AgencyID: g.agencyID, AgencyTimezone: g.agencyTZ}
	} else {
		agMap["AGENCY"] = agencyRec{AgencyID: g.agencyID, AgencyTimezone: g.agencyTZ}
	}
	sched["agency"] = agMap

	// routes
	routes := map[string]map[string]any{}
	for rid := range g.routes {
		routes[rid] = map[string]any{
			"route_id":         rid,
			"agency_id":        g.agencyID,
			"route_short_name": g.routeShortNames[rid],
		}
	}
	// If routes map is empty (e.g., loaded from ZIP without routes inserted into routes set), derive from routeShortNames
	if len(routes) == 0 {
		for rid := range g.routeShortNames {
			routes[rid] = map[string]any{
				"route_id":         rid,
				"agency_id":        g.agencyID,
				"route_short_name": g.routeShortNames[rid],
			}
		}
	}
	sched["routes"] = routes

	// stops
	stops := map[string]map[string]any{}
	for sid, name := range g.stopNames {
		coord := g.stopCoord[sid]
		stops[sid] = map[string]any{
			"stop_id":   sid,
			"stop_name": name,
			"stop_lat":  coord[1],
			"stop_lon":  coord[0],
		}
	}
	sched["stops"] = stops

	// trips
	trips := map[string]map[string]any{}
	for tripKey, routeID := range g.tripToRoute {
		dirStr := g.tripDirection[tripKey]
		dir := 0
		if dirStr == "1" {
			dir = 1
		}
		trips[tripKey] = map[string]any{
			"trip_id":       tripKey,
			"route_id":      routeID,
			"trip_headsign": g.tripHeadsign[tripKey],
			"direction_id":  dir,
			"shape_id":      g.tripShapeID[tripKey],
			"block_id":      g.tripBlockID[tripKey],
		}
	}
	sched["trips"] = trips

	return json.Marshal(sched)
}

// ExportIndexedSpatial emits a Node-like indexedSpatialData JSON blob.
func (g *GTFSIndex) ExportIndexedSpatial() ([]byte, error) {
	// paths: shapeID -> [{latitude, longitude, dist_traveled}]
	type pt struct {
		Latitude     float64 `json:"latitude"`
		Longitude    float64 `json:"longitude"`
		DistTraveled float64 `json:"dist_traveled"`
	}
	paths := map[string][]pt{}
	for shapeID, pts := range g.shapePoints {
		var arr []pt
		cum := g.shapeCumKM[shapeID]
		for i := range pts {
			d := 0.0
			if i < len(cum) {
				d = cum[i]
			}
			arr = append(arr, pt{Latitude: pts[i][1], Longitude: pts[i][0], DistTraveled: d})
		}
		paths[shapeID] = arr
	}
	spatial := map[string]any{
		"paths": paths,
	}
	return json.Marshal(spatial)
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
		for _, row := range rec[1:] {
			if rID >= 0 && rSN >= 0 {
				g.routeShortNames[row[rID]] = row[rSN]
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
		if tID < 0 || sID < 0 || sq < 0 {
			return nil
		}
		tmp := map[string][]struct {
			stop string
			seq  int
		}{}
		for _, row := range rec[1:] {
			trip := row[tID]
			stop := row[sID]
			seq, _ := strconv.Atoi(row[sq])
			tmp[trip] = append(tmp[trip], struct {
				stop string
				seq  int
			}{stop, seq})
		}
		for trip, arr := range tmp {
			sort.Slice(arr, func(i, j int) bool { return arr[i].seq < arr[j].seq })
			// first/last
			if len(arr) > 0 {
				g.tripOriginStop[trip] = arr[0].stop
				g.tripDestStop[trip] = arr[len(arr)-1].stop
			}
			// stop sequence + index map
			seqStops := make([]string, 0, len(arr))
			idxMap := make(map[string]int, len(arr))
			for i, v := range arr {
				seqStops = append(seqStops, v.stop)
				if _, ok := idxMap[v.stop]; !ok {
					idxMap[v.stop] = i
				}
			}
			g.tripStopSeq[trip] = seqStops
			g.tripStopIdx[trip] = idxMap
		}
	case "agency.txt":
		agID := idx("agency_id")
		agTZ := idx("agency_timezone")
		if len(rec) > 1 {
			if agID >= 0 && g.agencyID == "" {
				g.agencyID = rec[1][agID]
			}
			if agTZ >= 0 {
				g.agencyTZ = rec[1][agTZ]
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

// LoadZipAndExport runs the indexer: load a local GTFS zip and export Node-compatible JSON assets.
func (g *GTFSIndex) LoadZipAndExport(zipPath, outSchedulePath, outSpatialPath string) error {
	if zipPath == "" || outSchedulePath == "" || outSpatialPath == "" {
		return fmt.Errorf("zip and output paths required")
	}
	// reset maps
	*g = GTFSIndex{
		routeShortNames: map[string]string{},
		routes:          map[string]struct{}{},
		tripToRoute:     map[string]string{},
		tripHeadsign:    map[string]string{},
		tripOriginStop:  map[string]string{},
		tripDestStop:    map[string]string{},
		tripDirection:   map[string]string{},
		tripShapeID:     map[string]string{},
		tripBlockID:     map[string]string{},
		tripStopSeq:     map[string][]string{},
		tripStopIdx:     map[string]map[string]int{},
		stopNames:       map[string]string{},
		stopCoord:       map[string][2]float64{},
		shapePoints:     map[string][][2]float64{},
		shapeCumKM:      map[string][]float64{},
	}
	if err := g.loadFromLocalZip(zipPath); err != nil {
		return err
	}
	sched, err := g.ExportIndexedSchedule()
	if err != nil {
		return err
	}
	if err := os.WriteFile(outSchedulePath, sched, 0644); err != nil {
		return err
	}
	spatial, err := g.ExportIndexedSpatial()
	if err != nil {
		return err
	}
	if err := os.WriteFile(outSpatialPath, spatial, 0644); err != nil {
		return err
	}
	return nil
}

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
