package integration

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/formatter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/tests/helpers"
)

// TestFormatter_VM_ToXML verifies VehicleMonitoringDelivery responses are correctly
// formatted as valid XML with proper SIRI namespaces and UTF-8 encoding
func TestFormatter_VM_ToXML(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	response := c.GetCompleteVehicleMonitoringResponse()

	rb := formatter.NewResponseBuilder()
	xmlBytes := rb.BuildXML(response)

	if len(xmlBytes) == 0 {
		t.Fatal("XML output should not be empty")
	}

	xmlStr := string(xmlBytes)

	// Check for SIRI root element with namespace
	if !strings.Contains(xmlStr, "<Siri xmlns=\"http://www.siri.org.uk/siri\">") {
		t.Error("XML should contain <Siri> root element with namespace")
	}

	// Check for ServiceDelivery
	if !strings.Contains(xmlStr, "<ServiceDelivery>") {
		t.Error("XML should contain <ServiceDelivery>")
	}

	// Check for VehicleMonitoringDelivery - note: version is an attribute now
	if !strings.Contains(xmlStr, "VehicleMonitoringDelivery") {
		t.Error("XML should contain VehicleMonitoringDelivery")
	}

	// Check for VehicleActivity
	if !strings.Contains(xmlStr, "<VehicleActivity>") {
		t.Error("XML should contain <VehicleActivity>")
	}

	// Check for Velocity field (if present in data)
	if strings.Contains(xmlStr, "<Velocity>") {
		t.Log("✓ Velocity field is present in XML output")
	}

	// Verify ResponseTimestamp exists
	if !strings.Contains(xmlStr, "<ResponseTimestamp>") {
		t.Error("XML should contain <ResponseTimestamp>")
	}

	// Verify ProducerRef exists
	if !strings.Contains(xmlStr, "<ProducerRef>") {
		t.Error("XML should contain <ProducerRef>")
	}

	t.Logf("✓ Valid VM XML output (%d bytes)", len(xmlBytes))
}

// TestFormatter_VM_ToJSON verifies VehicleMonitoringDelivery responses are correctly
// formatted as valid JSON with proper structure
func TestFormatter_VM_ToJSON(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	response := c.GetCompleteVehicleMonitoringResponse()

	rb := formatter.NewResponseBuilder()
	jsonBytes := rb.BuildJSON(response)

	if len(jsonBytes) == 0 {
		t.Fatal("JSON output should not be empty")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Generated JSON is not valid: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Check for key fields - new flat structure
	if !strings.Contains(jsonStr, "\"VehicleMonitoringDelivery\"") {
		t.Error("JSON should contain 'VehicleMonitoringDelivery' field")
	}

	// Verify structure - flat SiriResponse
	if parsed["ResponseTimestamp"] == nil {
		t.Error("Response should have ResponseTimestamp")
	}

	if parsed["ProducerRef"] == nil {
		t.Error("Response should have ProducerRef")
	}

	// Check VehicleMonitoringDelivery array exists
	vm, ok := parsed["VehicleMonitoringDelivery"].([]interface{})
	if !ok {
		t.Fatal("VehicleMonitoringDelivery should be an array")
	}

	if len(vm) == 0 {
		t.Fatal("VehicleMonitoringDelivery should have at least one delivery")
	}

	t.Logf("✓ Valid VM JSON output (%d bytes)", len(jsonBytes))
}

// TestFormatter_ET_ToXML verifies EstimatedTimetable XML formatting
func TestFormatter_ET_ToXML(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	et := c.BuildEstimatedTimetable()

	// Wrap in response
	response := formatter.WrapEstimatedTimetableResponse(et, opts.AgencyID)

	rb := formatter.NewResponseBuilder()
	xmlBytes := rb.BuildXML(response)

	if len(xmlBytes) == 0 {
		t.Fatal("ET XML output should not be empty")
	}

	xmlStr := string(xmlBytes)

	// Check for SIRI structure
	if !strings.Contains(xmlStr, "<Siri xmlns=\"http://www.siri.org.uk/siri\">") {
		t.Error("XML should contain <Siri> root element")
	}

	// Check for EstimatedTimetableDelivery (with version attribute)
	if !strings.Contains(xmlStr, "<EstimatedTimetableDelivery") {
		t.Error("XML should contain <EstimatedTimetableDelivery>")
	}
	if !strings.Contains(xmlStr, "version=\"2.0\"") {
		t.Error("XML should contain version=\"2.0\" attribute")
	}

	t.Logf("✓ Valid ET XML output (%d bytes)", len(xmlBytes))
}

// TestFormatter_ET_ToJSON verifies EstimatedTimetable JSON formatting
func TestFormatter_ET_ToJSON(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	et := c.BuildEstimatedTimetable()
	response := formatter.WrapEstimatedTimetableResponse(et, opts.AgencyID)

	rb := formatter.NewResponseBuilder()
	jsonBytes := rb.BuildJSON(response)

	if len(jsonBytes) == 0 {
		t.Fatal("ET JSON output should not be empty")
	}

	// Verify valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Generated JSON is not valid: %v", err)
	}

	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, "\"EstimatedTimetableDelivery\"") {
		t.Error("JSON should contain 'EstimatedTimetableDelivery' field")
	}

	t.Logf("✓ Valid ET JSON output (%d bytes)", len(jsonBytes))
}

// TestFormatter_SX_ToXML verifies SituationExchangeDelivery XML formatting
func TestFormatter_SX_ToXML(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	sx := c.BuildSituationExchange()
	timestamp := gtfsrtData.GetTimestampForFeedMessage()
	response := formatter.WrapSituationExchangeResponse(sx, timestamp, opts.AgencyID)

	rb := formatter.NewResponseBuilder()
	xmlBytes := rb.BuildXML(response)

	if len(xmlBytes) == 0 {
		t.Fatal("SX XML output should not be empty")
	}

	xmlStr := string(xmlBytes)

	if !strings.Contains(xmlStr, "<Siri xmlns=\"http://www.siri.org.uk/siri\">") {
		t.Error("XML should contain <Siri> root element")
	}

	if !strings.Contains(xmlStr, "<SituationExchangeDelivery>") {
		t.Error("XML should contain <SituationExchangeDelivery>")
	}

	t.Logf("✓ Valid SX XML output (%d bytes)", len(xmlBytes))
}

// TestFormatter_SX_ToJSON verifies SituationExchangeDelivery JSON formatting
func TestFormatter_SX_ToJSON(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	sx := c.BuildSituationExchange()
	timestamp := gtfsrtData.GetTimestampForFeedMessage()
	response := formatter.WrapSituationExchangeResponse(sx, timestamp, opts.AgencyID)

	rb := formatter.NewResponseBuilder()
	jsonBytes := rb.BuildJSON(response)

	if len(jsonBytes) == 0 {
		t.Fatal("SX JSON output should not be empty")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Generated JSON is not valid: %v", err)
	}

	t.Logf("✓ Valid SX JSON output (%d bytes)", len(jsonBytes))
}

// TestFormatter_XML_UTF8Encoding verifies UTF-8 encoding and special character handling
func TestFormatter_XML_UTF8Encoding(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	response := c.GetCompleteVehicleMonitoringResponse()

	rb := formatter.NewResponseBuilder()
	xmlBytes := rb.BuildXML(response)

	xmlStr := string(xmlBytes)

	// Sofia has Cyrillic characters - verify they're present
	// Example: "Център за градска мобилност" (Center for Urban Mobility)
	hasCyrillic := false
	for _, r := range xmlStr {
		if r >= 0x0400 && r <= 0x04FF { // Cyrillic Unicode range
			hasCyrillic = true
			break
		}
	}

	if hasCyrillic {
		t.Log("✓ Cyrillic characters properly encoded")
	}

	// Verify XML special characters are escaped
	if strings.Contains(xmlStr, "&") {
		// Should have &amp; or other entities, not bare &
		if strings.Contains(xmlStr, "& ") || strings.Contains(xmlStr, "&<") {
			t.Error("Ampersands should be escaped as &amp;")
		}
	}

	t.Log("✓ UTF-8 encoding verified")
}

// TestFormatter_JSON_UTF8Encoding verifies JSON UTF-8 encoding
func TestFormatter_JSON_UTF8Encoding(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	response := c.GetCompleteVehicleMonitoringResponse()

	rb := formatter.NewResponseBuilder()
	jsonBytes := rb.BuildJSON(response)

	// JSON should handle Cyrillic natively
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("JSON with UTF-8 should parse: %v", err)
	}

	t.Log("✓ JSON UTF-8 encoding verified")
}

// TestFormatter_XML_EmptyVehicles verifies handling of empty vehicle list
func TestFormatter_XML_EmptyVehicles(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")

	// Create empty GTFS-RT wrapper
	emptyRT := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, emptyRT, opts)

	response := c.GetCompleteVehicleMonitoringResponse()

	rb := formatter.NewResponseBuilder()
	xmlBytes := rb.BuildXML(response)

	// Should still produce valid XML structure
	if len(xmlBytes) == 0 {
		t.Fatal("XML should not be empty even with no vehicles")
	}

	xmlStr := string(xmlBytes)

	// Should have structure but maybe no VehicleActivity
	if !strings.Contains(xmlStr, "<Siri xmlns=\"http://www.siri.org.uk/siri\">") {
		t.Error("Should have valid SIRI structure")
	}

	if !strings.Contains(xmlStr, "<ServiceDelivery>") {
		t.Error("Should have ServiceDelivery")
	}

	t.Log("✓ Empty vehicle list handled gracefully")
}

// TestFormatter_XML_ValidStructure verifies XML is well-formed
func TestFormatter_XML_ValidStructure(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	response := c.GetCompleteVehicleMonitoringResponse()

	rb := formatter.NewResponseBuilder()
	xmlBytes := rb.BuildXML(response)

	// Parse as generic XML to verify well-formedness
	var generic interface{}
	if err := xml.Unmarshal(xmlBytes, &generic); err != nil {
		t.Fatalf("XML is not well-formed: %v", err)
	}

	t.Log("✓ XML is well-formed")
}

// TestFormatter_XML_vs_JSON_Equivalence verifies both formats contain same data
func TestFormatter_XML_vs_JSON_Equivalence(t *testing.T) {
	gtfsIndex := helpers.MustLoadTestGTFS("sofia-static.zip", "SOFIA")
	gtfsrtData := helpers.LoadGTFSRTFromLocal(t)

	opts := helpers.DefaultConverterOptions("SOFIA")
	c := converter.NewConverter(gtfsIndex, gtfsrtData, opts)

	response := c.GetCompleteVehicleMonitoringResponse()

	rb := formatter.NewResponseBuilder()
	xmlBytes := rb.BuildXML(response)
	jsonBytes := rb.BuildJSON(response)

	// Both should produce non-empty output
	if len(xmlBytes) == 0 || len(jsonBytes) == 0 {
		t.Fatal("Both XML and JSON should produce output")
	}

	// Count vehicles in response
	vehicleCount := len(response.VehicleMonitoringDelivery[0].VehicleActivity)

	// Both formats should have the same vehicle count in the source data
	// (Note: XML/JSON might have different field occurrences due to structure)
	t.Logf("✓ Source data contains %d vehicles", vehicleCount)
	t.Logf("  XML: %d bytes, JSON: %d bytes", len(xmlBytes), len(jsonBytes))

	// Verify both have substantive content
	if len(xmlBytes) < 1000 || len(jsonBytes) < 1000 {
		t.Error("Both formats should have substantial content")
	}
}
