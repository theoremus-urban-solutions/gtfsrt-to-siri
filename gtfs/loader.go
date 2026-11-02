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
		TripStopTimes:   map[string][]StopTime{},
		StopNames:       map[string]string{},
		StopCoord:       map[string][2]float64{},
	}

	// Parse GTFS files from zip
	if err := index.parseGTFSFiles(zipReader); err != nil {
		return nil, fmt.Errorf("failed to parse GTFS: %w", err)
	}

	return index, nil
}

// parseGTFSFiles reads all required GTFS files from zip
func (g *GTFSIndex) parseGTFSFiles(zipReader *zip.Reader) error {
	in := interner{}
	for _, f := range zipReader.File {
		name := strings.ToLower(f.Name)
		if name == "routes.txt" || name == "trips.txt" || name == "stops.txt" ||
			name == "stop_times.txt" || name == "agency.txt" {
			if err := g.consumeCSV(f, in); err != nil {
				return err
			}
		}
	}

	return nil
}

// interner keeps one copy of each repeated value so kept strings do not pin whole CSV lines.
type interner map[string]string

func (in interner) get(v string) string {
	if s, ok := in[v]; ok {
		return s
	}
	s := strings.Clone(v)
	in[s] = s
	return s
}

func (g *GTFSIndex) consumeCSV(f *zip.File, in interner) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	csvr := csv.NewReader(r)
	csvr.LazyQuotes = true // Handle malformed CSV with quotes
	csvr.ReuseRecord = true
	first, err := csvr.Read()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	head := append([]string(nil), first...)
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
	// rows calls fn for each data row; the row slice is reused between calls.
	rows := func(fn func(row []string)) error {
		for {
			row, err := csvr.Read()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			fn(row)
		}
	}
	switch strings.ToLower(f.Name) {
	case "routes.txt":
		rID := idx("route_id")
		rSN := idx("route_short_name")
		rType := idx("route_type")
		return rows(func(row []string) {
			if rID >= 0 && rSN >= 0 {
				g.RouteShortNames[in.get(row[rID])] = in.get(row[rSN])
			}
			if rID >= 0 && rType >= 0 {
				if typeInt, err := strconv.Atoi(row[rType]); err == nil {
					g.RouteTypes[in.get(row[rID])] = typeInt
				}
			}
		})
	case "trips.txt":
		rID := idx("route_id")
		tID := idx("trip_id")
		hs := idx("trip_headsign")
		dir := idx("direction_id")
		blk := idx("block_id")
		return rows(func(row []string) {
			if tID < 0 {
				return
			}
			trip := in.get(row[tID])
			if rID >= 0 {
				g.TripToRoute[trip] = in.get(row[rID])
			}
			if hs >= 0 {
				g.TripHeadsign[trip] = in.get(row[hs])
			}
			if dir >= 0 {
				g.TripDirection[trip] = in.get(row[dir])
			}
			if blk >= 0 {
				g.TripBlockID[trip] = in.get(row[blk])
			}
		})
	case "stops.txt":
		sID := idx("stop_id")
		sN := idx("stop_name")
		sLat := idx("stop_lat")
		sLon := idx("stop_lon")
		return rows(func(row []string) {
			if sID >= 0 && sN >= 0 {
				g.StopNames[in.get(row[sID])] = in.get(row[sN])
			}
			if sID >= 0 && sLat >= 0 && sLon >= 0 {
				lat, _ := strconv.ParseFloat(row[sLat], 64)
				lon, _ := strconv.ParseFloat(row[sLon], 64)
				g.StopCoord[in.get(row[sID])] = [2]float64{lon, lat}
			}
		})
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
		type stopRow struct {
			stop string
			seq  int
			st   StopTime
		}
		tmp := map[string][]stopRow{}
		err := rows(func(row []string) {
			seq, _ := strconv.Atoi(row[sq])
			var st StopTime
			if arrTime >= 0 && arrTime < len(row) {
				st.ArrivalTime = in.get(row[arrTime])
			}
			if depTime >= 0 && depTime < len(row) {
				st.DepartureTime = in.get(row[depTime])
			}
			if pickupType >= 0 && pickupType < len(row) && row[pickupType] != "" {
				pickup, _ := strconv.Atoi(row[pickupType])
				st.PickupType = int8(pickup)
			}
			if dropOffType >= 0 && dropOffType < len(row) && row[dropOffType] != "" {
				dropOff, _ := strconv.Atoi(row[dropOffType])
				st.DropOffType = int8(dropOff)
			}
			trip := in.get(row[tID])
			tmp[trip] = append(tmp[trip], stopRow{stop: in.get(row[sID]), seq: seq, st: st})
		})
		if err != nil {
			return err
		}
		for trip, arr := range tmp {
			sort.Slice(arr, func(i, j int) bool { return arr[i].seq < arr[j].seq })
			if len(arr) > 0 {
				g.TripOriginStop[trip] = arr[0].stop
				g.TripDestStop[trip] = arr[len(arr)-1].stop
			}
			seqStops := make([]string, len(arr))
			times := make([]StopTime, len(arr))
			for i, v := range arr {
				seqStops[i] = v.stop
				times[i] = v.st
			}
			g.TripStopSeq[trip] = seqStops
			g.TripStopTimes[trip] = times
		}
	case "agency.txt":
		agID := idx("agency_id")
		agTZ := idx("agency_timezone")
		agName := idx("agency_name")
		row, err := csvr.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if agID >= 0 && g.AgencyID == "" {
			g.AgencyID = strings.Clone(row[agID])
		}
		if agTZ >= 0 {
			g.AgencyTZ = strings.Clone(row[agTZ])
		}
		if agName >= 0 {
			g.AgencyName = strings.Clone(row[agName])
		}
		return rows(func([]string) {})
	}
	return nil
}
