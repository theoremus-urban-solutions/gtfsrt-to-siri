package unit

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
	"time"

	gtfsrtpb "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/formatter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfs"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

func TestParseGTFSTimeInLocation(t *testing.T) {
	sofia, err := time.LoadLocation("Europe/Sofia")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, gtfsTime, date string
		want                 time.Time
	}{
		{"winter day", "14:40:00", "20261110", time.Date(2026, 11, 10, 12, 40, 0, 0, time.UTC)},
		{"past midnight", "25:30:00", "20261110", time.Date(2026, 11, 10, 23, 30, 0, 0, time.UTC)},
		// Clocks go from 03:00 to 04:00 that night; GTFS counts from noon minus 12h, so 08:00 is 08:00 EEST.
		{"daylight saving start", "08:00:00", "20260329", time.Date(2026, 3, 29, 5, 0, 0, 0, time.UTC)},
		{"summer day", "14:40:00", "20260710", time.Date(2026, 7, 10, 11, 40, 0, 0, time.UTC)},
	}
	for _, c := range cases {
		got := utils.ParseGTFSTimeInLocation(c.gtfsTime, c.date, sofia)
		if got != c.want.Unix() {
			t.Errorf("%s: got %s, want %s", c.name, time.Unix(got, 0).UTC(), c.want)
		}
	}
	if utils.ParseGTFSTimeInLocation("bad", "20261110", sofia) != 0 || utils.ParseGTFSTimeInLocation("08:00:00", "2026111", sofia) != 0 {
		t.Error("malformed input did not give 0")
	}
}

func sofiaGTFSZip(t *testing.T) []byte {
	t.Helper()
	files := map[string]string{
		"agency.txt":     "agency_id,agency_name,agency_url,agency_timezone\nA,Sofia,http://sofia.example,Europe/Sofia\n",
		"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nS1,One,42.70,23.32\nS2,Two,42.71,23.33\n",
		"routes.txt":     "route_id,agency_id,route_short_name,route_long_name,route_type\nR,A,1,One - Two,3\n",
		"trips.txt":      "route_id,service_id,trip_id\nR,D,T\n",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT,14:30:00,14:30:00,S1,1\nT,14:40:00,14:40:00,S2,2\n",
		"calendar.txt":   "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nD,1,1,1,1,1,1,1,20260101,20271231\n",
	}
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for name, body := range files {
		f, _ := w.Create(name)
		_, _ = f.Write([]byte(body))
	}
	_ = w.Close()
	return buf.Bytes()
}

// TestEstimatedTimetableUsesAgencyTimezone: a 14:40 Sofia stop is aimed at 12:40 UTC even when the process runs in UTC.
func TestEstimatedTimetableUsesAgencyTimezone(t *testing.T) {
	prev := time.Local
	time.Local = time.UTC
	defer func() { time.Local = prev }()

	idx, err := gtfs.NewGTFSIndexFromBytes(sofiaGTFSZip(t), "SOF")
	if err != nil {
		t.Fatal(err)
	}
	expected := time.Date(2026, 11, 10, 12, 42, 0, 0, time.UTC).Unix()
	tu, _ := proto.Marshal(&gtfsrtpb.FeedMessage{
		Header: &gtfsrtpb.FeedHeader{GtfsRealtimeVersion: proto.String("2.0"), Timestamp: proto.Uint64(uint64(expected - 600))},
		Entity: []*gtfsrtpb.FeedEntity{{Id: proto.String("e"), TripUpdate: &gtfsrtpb.TripUpdate{
			Trip: &gtfsrtpb.TripDescriptor{TripId: proto.String("T"), RouteId: proto.String("R"), StartDate: proto.String("20261110")},
			StopTimeUpdate: []*gtfsrtpb.TripUpdate_StopTimeUpdate{{StopId: proto.String("S2"), StopSequence: proto.Uint32(2),
				Arrival: &gtfsrtpb.TripUpdate_StopTimeEvent{Time: proto.Int64(expected)}}},
		}}},
	})
	rt, err := gtfsrt.NewGTFSRTWrapper(tu, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	conv := converter.NewConverterWithCachedGTFS(idx, rt, converter.ConverterOptions{AgencyID: "SOF"})
	xml := string(formatter.NewResponseBuilder().BuildXML(formatter.WrapEstimatedTimetableResponse(conv.BuildEstimatedTimetable(), "SOF")))
	if !strings.Contains(xml, "<AimedArrivalTime>2026-11-10T12:40:00") {
		t.Fatalf("aimed arrival is not 12:40 UTC:\n%s", xml)
	}
	if !strings.Contains(xml, "<ArrivalStatus>delayed</ArrivalStatus>") {
		t.Fatalf("a 2 minute late arrival is not delayed:\n%s", xml)
	}
}
