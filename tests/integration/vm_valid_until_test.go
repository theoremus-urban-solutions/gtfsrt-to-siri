package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/formatter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/tests/helpers"
)

// TestVM_ValidUntilTime checks that every VehicleActivity carries the ValidUntilTime SIRI requires.
func TestVM_ValidUntilTime(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	rt := helpers.LoadGTFSRTFromLocal(t)

	for _, tc := range []struct {
		name           string
		readIntervalMS int64
		want           time.Duration
	}{
		{"default when zero", 0, 30 * time.Second},
		{"configured interval", 60000, 60 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := helpers.DefaultConverterOptions("SOFIA")
			opts.ReadIntervalMS = tc.readIntervalMS
			resp := converter.NewConverter(gtfsIndex, rt, opts).GetCompleteVehicleMonitoringResponse()

			checked := 0
			for _, va := range resp.VehicleMonitoringDelivery[0].VehicleActivity {
				recorded, err := time.Parse(time.RFC3339, va.RecordedAtTime)
				if err != nil || recorded.Unix() <= 0 {
					continue
				}
				validUntil, err := time.Parse(time.RFC3339, va.ValidUntilTime)
				if err != nil {
					t.Fatalf("ValidUntilTime %q for RecordedAtTime %q: %v", va.ValidUntilTime, va.RecordedAtTime, err)
				}
				if got := validUntil.Sub(recorded); got != tc.want {
					t.Fatalf("ValidUntilTime - RecordedAtTime = %v, want %v", got, tc.want)
				}
				checked++
			}
			if checked == 0 {
				t.Fatal("no VehicleActivity with a RecordedAtTime to check")
			}

			xml := string(formatter.NewResponseBuilder().BuildXML(resp))
			if n, m := strings.Count(xml, "<VehicleActivity>"), strings.Count(xml, "</RecordedAtTime><ValidUntilTime>"); n != m {
				t.Fatalf("%d VehicleActivity elements but %d ValidUntilTime right after RecordedAtTime", n, m)
			}
		})
	}
}
