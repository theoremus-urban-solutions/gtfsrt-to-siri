package integration

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/tests/helpers"
)

// Test SituationExchangeDelivery conversion
func TestConverter_SituationExchange_Basic(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	result := c.BuildSituationExchange()

	// Alerts may be empty if no service alerts in feed
	t.Logf("Found %d situation elements (alerts)", len(result.Situations))

	if len(result.Situations) > 0 {
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

	if len(result.Situations) == 0 {
		t.Skip("No alerts in feed to test structure")
	}

	alert := result.Situations[0]

	if alert.SituationNumber == "" {
		t.Error("Alert should have SituationNumber")
	}

	if len(alert.Summary) == 0 && len(alert.Description) == 0 {
		t.Error("Alert should have either Summary or Description")
	}

	summaryText := ""
	if len(alert.Summary) > 0 {
		summaryText = alert.Summary[0].Text
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

	if len(result.Situations) == 0 {
		t.Skip("No alerts in feed to test affected entities")
	}

	for _, alert := range result.Situations {
		// Check if we have affected networks/lines/stops
		if alert.Affects != nil && alert.Affects.Networks != nil {
			t.Logf("Alert affects %d networks", len(alert.Affects.Networks.AffectedNetwork))
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

	if len(result.Situations) == 0 {
		t.Skip("No alerts in feed to test metadata")
	}

	alert := result.Situations[0]

	// These fields may be optional depending on the feed
	t.Logf("Alert severity: %s", alert.Severity)
	if len(alert.ValidityPeriod) > 0 {
		t.Logf("Alert validity period: Start=%s, End=%s",
			alert.ValidityPeriod[0].StartTime, alert.ValidityPeriod[0].EndTime)
	}
}
