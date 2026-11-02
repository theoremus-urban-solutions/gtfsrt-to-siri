package unit

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfs"
)

// loopGTFS builds a feed whose trip T1 visits stop A twice: A, B, A.
func loopGTFS(t *testing.T) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	files := map[string]string{
		"agency.txt":     "agency_id,agency_name,agency_timezone\nAG,Agency,Europe/Sofia\n",
		"routes.txt":     "route_id,route_short_name,route_type\nR1,1,3\n",
		"trips.txt":      "route_id,trip_id,trip_headsign,direction_id\nR1,T1,Loop,0\n",
		"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,42.1,23.1\nB,Stop B,42.2,23.2\n",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence,pickup_type,drop_off_type\nT1,10:20:00,10:21:00,A,3,0,1\nT1,10:00:00,10:01:00,A,1,1,0\nT1,10:10:00,10:11:00,B,2,0,0\n",
	}
	for name, body := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestGTFSIndex_LoopTripStopLookups(t *testing.T) {
	idx, err := gtfs.NewGTFSIndexFromBytes(loopGTFS(t), "AG")
	if err != nil {
		t.Fatal(err)
	}
	if got := idx.TripStopSeq["T1"]; len(got) != 3 || got[0] != "A" || got[1] != "B" || got[2] != "A" {
		t.Fatalf("stop sequence = %v, want [A B A]", got)
	}
	// Times and pickup/drop-off come from the last visit of a stop.
	if got := idx.GetArrivalTime("T1", "A"); got != "10:20:00" {
		t.Errorf("arrival at A = %q, want 10:20:00", got)
	}
	if got := idx.GetDepartureTime("T1", "A"); got != "10:21:00" {
		t.Errorf("departure at A = %q, want 10:21:00", got)
	}
	if idx.GetPickupType("T1", "A") != 0 || idx.GetDropOffType("T1", "A") != 1 {
		t.Errorf("pickup/drop-off at A = %d/%d, want 0/1", idx.GetPickupType("T1", "A"), idx.GetDropOffType("T1", "A"))
	}
	// The previous stop comes from the first visit of a stop.
	if got := idx.GetPreviousStopIDOfStopForTrip("T1", "A"); got != "" {
		t.Errorf("previous of A = %q, want empty", got)
	}
	if got := idx.GetPreviousStopIDOfStopForTrip("T1", "B"); got != "A" {
		t.Errorf("previous of B = %q, want A", got)
	}
	if got := idx.GetArrivalTime("T1", "C"); got != "" {
		t.Errorf("arrival at unknown stop = %q, want empty", got)
	}
	if got := idx.GetArrivalTime("T9", "A"); got != "" {
		t.Errorf("arrival on unknown trip = %q, want empty", got)
	}
}

func TestGTFSIndex_SerializeRoundTripKeepsStopTimes(t *testing.T) {
	idx, err := gtfs.NewGTFSIndexFromBytes(loopGTFS(t), "AG")
	if err != nil {
		t.Fatal(err)
	}
	data, err := gtfs.SerializeIndex(idx)
	if err != nil {
		t.Fatal(err)
	}
	back, err := gtfs.DeserializeIndex(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := back.GetArrivalTime("T1", "B"); got != "10:10:00" {
		t.Errorf("arrival at B after round trip = %q, want 10:10:00", got)
	}
}

func TestGTFSIndex_DeserializeRejectsIndexWithoutStopTimes(t *testing.T) {
	data, err := gtfs.SerializeIndex(&gtfs.GTFSIndex{TripStopSeq: map[string][]string{"T1": {"A"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gtfs.DeserializeIndex(data); err == nil {
		t.Error("expected an error for an index with stop sequences but no stop times")
	}
}
