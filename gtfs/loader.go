package gtfs

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

// FetchGTFSData fetches GTFS data from a URL or file path and returns raw bytes.
// This is a CLI helper function - library users should fetch data themselves.
func FetchGTFSData(urlOrPath string) ([]byte, error) {
	// Check if it's a file path
	if info, err := os.Stat(urlOrPath); err == nil && !info.IsDir() {
		return os.ReadFile(urlOrPath)
	}

	// Treat as URL
	resp, err := http.Get(urlOrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GTFS from %s: %w", urlOrPath, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d fetching GTFS from %s", resp.StatusCode, urlOrPath)
	}

	return io.ReadAll(resp.Body)
}

// NewGTFSIndexFromBytes creates a GTFS index from raw zip bytes.
// This is the primary constructor for data-source agnostic loading.
//
// Example:
//
//	zipBytes := fetchGTFSFromYourSource() // HTTP, MinIO, files, etc.
//	index, err := gtfs.NewGTFSIndexFromBytes(zipBytes, "AGENCY_ID")
func NewGTFSIndexFromBytes(zipData []byte, agencyID string) (*GTFSIndex, error) {
	return NewGTFSIndexFromReader(bytes.NewReader(zipData), int64(len(zipData)), agencyID)
}

// NewGTFSIndexFromReader creates a GTFS index from an io.ReaderAt.
// Use this for streaming or when you already have an open reader.
//
// Example:
//
//	file, _ := os.Open("gtfs.zip")
//	defer file.Close()
//	stat, _ := file.Stat()
//	index, err := gtfs.NewGTFSIndexFromReader(file, stat.Size(), "AGENCY_ID")
func NewGTFSIndexFromReader(r io.ReaderAt, size int64, agencyID string) (*GTFSIndex, error) {
	// Create zip reader from ReaderAt
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("failed to open GTFS zip: %w", err)
	}

	// Initialize index
	index := &GTFSIndex{
		AgencyID:        agencyID,
		RouteShortNames: map[string]string{},
		RouteTypes:      map[string]int{},
		Routes:          map[string]struct{}{},
		TripToRoute:     map[string]string{},
		TripHeadsign:    map[string]string{},
		TripOriginStop:  map[string]string{},
		TripDestStop:    map[string]string{},
		TripDirection:   map[string]string{},
		TripBlockID:     map[string]string{},
		TripStopSeq:     map[string][]string{},
		TripStopIdx:     map[string]map[string]int{},
		StopNames:       map[string]string{},
		StopCoord:       map[string][2]float64{},
		StopTimes:       map[string]map[string]StopTime{},
	}

	// Parse GTFS files from zip
	if err := index.parseGTFSFiles(zipReader); err != nil {
		return nil, fmt.Errorf("failed to parse GTFS: %w", err)
	}

	return index, nil
}

// parseGTFSFiles reads all required GTFS files from zip
func (g *GTFSIndex) parseGTFSFiles(zipReader *zip.Reader) error {
	for _, f := range zipReader.File {
		name := strings.ToLower(f.Name)
		if name == "routes.txt" || name == "trips.txt" || name == "stops.txt" ||
			name == "stop_times.txt" || name == "agency.txt" {
			if err := g.consumeCSV(f); err != nil {
				return err
			}
		}
	}

	return nil
}

// Utility converters for flexible JSON values

func (g *GTFSIndex) consumeCSV(f *zip.File) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	csvr := csv.NewReader(r)
	csvr.LazyQuotes = true // Handle malformed CSV with quotes
	rec, err := csvr.ReadAll()
	if err != nil {
		return err
	}
	if len(rec) == 0 {
		return nil
	}
	head := rec[0]
	// Strip UTF-8 BOM from first column header if present
	if len(head) > 0 {
		head[0] = strings.TrimPrefix(head[0], "\ufeff")
	}
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
				g.RouteShortNames[row[rID]] = row[rSN]
			}
			if rID >= 0 && rType >= 0 {
				if typeInt, err := strconv.Atoi(row[rType]); err == nil {
					g.RouteTypes[row[rID]] = typeInt
				}
			}
		}
	case "trips.txt":
		rID := idx("route_id")
		tID := idx("trip_id")
		hs := idx("trip_headsign")
		dir := idx("direction_id")
		blk := idx("block_id")
		for _, row := range rec[1:] {
			if tID >= 0 && rID >= 0 {
				g.TripToRoute[row[tID]] = row[rID]
			}
			if tID >= 0 && hs >= 0 {
				g.TripHeadsign[row[tID]] = row[hs]
			}
			if tID >= 0 && dir >= 0 {
				g.TripDirection[row[tID]] = row[dir]
			}
			if tID >= 0 && blk >= 0 {
				g.TripBlockID[row[tID]] = row[blk]
			}
		}
	case "stops.txt":
		sID := idx("stop_id")
		sN := idx("stop_name")
		sLat := idx("stop_lat")
		sLon := idx("stop_lon")
		for _, row := range rec[1:] {
			if sID >= 0 && sN >= 0 {
				g.StopNames[row[sID]] = row[sN]
			}
			if sID >= 0 && sLat >= 0 && sLon >= 0 {
				lat, _ := strconv.ParseFloat(row[sLat], 64)
				lon, _ := strconv.ParseFloat(row[sLon], 64)
				g.StopCoord[row[sID]] = [2]float64{lon, lat}
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
				g.TripOriginStop[trip] = arr[0].stop
				g.TripDestStop[trip] = arr[len(arr)-1].stop
			}
			// Initialize map for this trip
			g.StopTimes[trip] = make(map[string]StopTime)
			// stop sequence + index map + stop times data
			seqStops := make([]string, 0, len(arr))
			idxMap := make(map[string]int, len(arr))
			for i, v := range arr {
				seqStops = append(seqStops, v.stop)
				if _, ok := idxMap[v.stop]; !ok {
					idxMap[v.stop] = i
				}
				// Store stop time data in consolidated struct
				g.StopTimes[trip][v.stop] = StopTime{
					ArrivalTime:   v.arrTime,
					DepartureTime: v.depTime,
					PickupType:    int8(v.pickupType),
					DropOffType:   int8(v.dropOffType),
				}
			}
			g.TripStopSeq[trip] = seqStops
			g.TripStopIdx[trip] = idxMap
		}
	case "agency.txt":
		agID := idx("agency_id")
		agTZ := idx("agency_timezone")
		agName := idx("agency_name")
		if len(rec) > 1 {
			if agID >= 0 && g.AgencyID == "" {
				g.AgencyID = rec[1][agID]
			}
			if agTZ >= 0 {
				g.AgencyTZ = rec[1][agTZ]
			}
			if agName >= 0 {
				g.AgencyName = rec[1][agName]
			}
		}
	}
	return nil
}
