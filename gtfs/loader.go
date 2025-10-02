package gtfs

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

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
	for shapeID, pts := range g.ShapePoints {
		g.ShapeCumKM[shapeID] = cumulativeKM(pts)
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
	for shapeID, pts := range g.ShapePoints {
		g.ShapeCumKM[shapeID] = cumulativeKM(pts)
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
			g.TripStopSeq[trip] = seqStops
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
			g.ShapePoints[shapeID] = pts
		}
	}
	return nil
}

// dumpDebugJSON is optionally used for debugging loaded data
func (g *GTFSIndex) dumpDebugJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
