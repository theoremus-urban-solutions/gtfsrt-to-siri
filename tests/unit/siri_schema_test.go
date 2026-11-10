package unit

import (
	"regexp"
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/formatter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/tests/helpers"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
	"github.com/theoremus-urban-solutions/transit-types/siri"
)

// These tests pin the SIRI 2.1 XSD rules the Sofia fixtures used to break; each failed xmllint before.

func sofiaXML(t *testing.T) (vm, et, sx string) {
	t.Helper()
	conv := helpers.CreateTestConverter(t, "BG", helpers.DefaultConverterOptions("BG"))
	rt := helpers.LoadTestGTFSRT(t)
	rb := formatter.NewResponseBuilder()
	vm = string(rb.BuildXML(conv.GetCompleteVehicleMonitoringResponse()))
	et = string(rb.BuildXML(formatter.WrapEstimatedTimetableResponse(conv.BuildEstimatedTimetable(), "BG")))
	sx = string(rb.BuildXML(formatter.WrapSituationExchangeResponse(conv.BuildSituationExchange(), rt.GetTimestampForFeedMessage(), "BG")))
	return vm, et, sx
}

func values(xml, element string) []string {
	var out []string
	for _, m := range regexp.MustCompile("<"+element+">([^<]*)</"+element+">").FindAllStringSubmatch(xml, -1) {
		out = append(out, m[1])
	}
	return out
}

func TestSIRISchemaRules(t *testing.T) {
	vm, et, sx := sofiaXML(t)
	modes := map[string]bool{"air": true, "bus": true, "coach": true, "ferry": true, "metro": true, "rail": true, "tram": true, "underground": true}
	for name, doc := range map[string]string{"VM": vm, "ET": et} {
		ops := values(doc, "OperatorRef")
		if len(ops) == 0 {
			t.Fatalf("%s has no OperatorRef", name)
		}
		for _, op := range ops {
			if op != "BG:Operator:A" {
				t.Fatalf("%s OperatorRef %q, want agency.txt's agency_id as BG:Operator:A", name, op)
			}
		}
		for _, m := range values(doc, "VehicleMode") {
			if !modes[m] {
				t.Fatalf("%s VehicleMode %q is not a SIRI 2.1 mode", name, m)
			}
		}
		for _, el := range []string{"OriginName", "DestinationName"} {
			for _, v := range values(doc, el) {
				if strings.ContainsAny(v, ",[]{}?$%^=@#;:") || v == "" {
					t.Fatalf("%s %s %q holds a character SIRI forbids", name, el, v)
				}
			}
		}
	}
	// The fixture trips carry no start_date, so every DataFrameRef comes from the feed's service day.
	if n, refs := strings.Count(vm, "<FramedVehicleJourneyRef>"), strings.Count(vm, "<DataFrameRef>"); n == 0 || n != refs {
		t.Fatalf("VM has %d FramedVehicleJourneyRef and %d DataFrameRef", n, refs)
	}
	if strings.Contains(vm, "</Delay><InCongestion>") {
		t.Fatal("VM writes InCongestion after Delay; SIRI wants it right after Monitored")
	}
	conds := map[string]bool{"noService": true, "limitedOperation": true, "majorDelays": true, "diverted": true, "additionalService": true, "altered": true, "stopMoved": true}
	cs := values(sx, "Condition")
	if len(cs) == 0 {
		t.Fatal("SX fixture gave no Consequence")
	}
	for _, c := range cs {
		if !conds[c] {
			t.Fatalf("SX Condition %q is not a SIRI 2.1 ServiceCondition", c)
		}
	}
	for _, el := range strings.Split(sx, "<PtSituationElement>")[1:] {
		if i, a := strings.Index(el, "<InfoLinks>"), strings.Index(el, "<Affects>"); i >= 0 && a >= 0 && i > a {
			t.Fatal("SX writes InfoLinks after Affects")
		}
	}
	if !strings.Contains(sx, "<SituationExchangeDelivery") || !strings.Contains(sx, "<ResponseTimestamp>") {
		t.Fatal("SX delivery lacks its ResponseTimestamp")
	}
}

func TestEmptySituationExchangeKeepsDelivery(t *testing.T) {
	xml := string(formatter.NewResponseBuilder().BuildXML(formatter.WrapSituationExchangeResponse(siri.SituationExchangeDelivery{}, 1794304800, "BG")))
	want := "<SituationExchangeDelivery><ResponseTimestamp>"
	if !strings.Contains(xml, want) || strings.Contains(xml, "<Situations>") {
		t.Fatalf("empty SX is not a delivery with only its timestamp:\n%s", xml)
	}
}

func TestSIRINames(t *testing.T) {
	if got := utils.NMToken("BG:Operator:Център за градска мобилност"); got != "BG:Operator:Център_за_градска_мобилност" {
		t.Errorf("NMToken: %q", got)
	}
	if got := utils.PlaceName("Ж.К. ЛЮЛИН-1,2"); got != "Ж.К. ЛЮЛИН-1 2" {
		t.Errorf("PlaceName: %q", got)
	}
	if got := utils.PlaceName(" ; "); got != "" {
		t.Errorf("PlaceName of punctuation only: %q", got)
	}
}
