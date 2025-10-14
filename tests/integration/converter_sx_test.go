package integration

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/tests/helpers"
)

// Test SituationExchangeDelivery conversion
func TestConverter_SituationExchange_Basic(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	result := c.BuildSituationExchange()

	// Extract situations from the any type
	situations, ok := result.Situations.([]siri.PtSituationElement)
	if !ok {
		t.Fatal("Situations should be []PtSituationElement")
	}

	// Alerts may be empty if no service alerts in feed
	t.Logf("Found %d situation elements (alerts)", len(situations))

	if len(situations) > 0 {
		t.Logf("✓ Alerts present in feed")
	} else {
		t.Log("No alerts in current Sofia feed (this is normal)")
	}
}

// Test SX alert structure if alerts exist
func TestConverter_SX_AlertStructure(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	result := c.BuildSituationExchange()

	situations, ok := result.Situations.([]siri.PtSituationElement)
	if !ok {
		t.Fatal("Situations should be []PtSituationElement")
	}

	if len(situations) == 0 {
		t.Skip("No alerts in feed to test structure")
	}

	alert := situations[0]

	if alert.SituationNumber == "" {
		t.Error("Alert should have SituationNumber")
	}

	if len(alert.Summaries) == 0 && len(alert.Descriptions) == 0 {
		t.Error("Alert should have either Summaries or Descriptions")
	}

	summaryText := ""
	if len(alert.Summaries) > 0 {
		summaryText = alert.Summaries[0].Text
	}
	t.Logf("Alert: %s - %s", alert.SituationNumber, summaryText)
}

// Test SX affected entities (routes, stops, trips)
func TestConverter_SX_AffectedEntities(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	result := c.BuildSituationExchange()

	situations, ok := result.Situations.([]siri.PtSituationElement)
	if !ok {
		t.Fatal("Situations should be []PtSituationElement")
	}

	if len(situations) == 0 {
		t.Skip("No alerts in feed to test affected entities")
	}

	for _, alert := range situations {
		// Check if we have affected networks/lines/stops
		if len(alert.Affects.Networks) > 0 {
			t.Logf("Alert affects %d networks", len(alert.Affects.Networks))
		}
		// Can add more affected entity checks here
	}
}

// Test SX severity and validity period
func TestConverter_SX_MetadataFields(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	result := c.BuildSituationExchange()

	situations, ok := result.Situations.([]siri.PtSituationElement)
	if !ok {
		t.Fatal("Situations should be []PtSituationElement")
	}

	if len(situations) == 0 {
		t.Skip("No alerts in feed to test metadata")
	}

	alert := situations[0]

	// These fields may be optional depending on the feed
	t.Logf("Alert severity: %s", alert.Severity)
	t.Logf("Alert publication window: Start=%s, End=%s",
		alert.PublicationWindow.StartTime, alert.PublicationWindow.EndTime)
}
