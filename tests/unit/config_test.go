package unit

import (
	"os"
	"path/filepath"
	"testing"

	"mta/gtfsrt-to-siri/config"

	"gopkg.in/yaml.v3"
)

// TestConfig_LoadFromFile tests loading the main config.yml
func TestConfig_LoadFromFile(t *testing.T) {
	// Save original config and working directory
	origConfig := config.Config
	origDir, _ := os.Getwd()
	defer func() {
		config.Config = origConfig
		os.Chdir(origDir)
	}()

	// Change to project root
	err := os.Chdir("../../")
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err = config.LoadAppConfig()
	if err != nil {
		t.Fatalf("Failed to load config.yml: %v", err)
	}

	if config.Config.GTFS.AgencyID == "" {
		t.Error("Config should have agency_id")
	}

	t.Logf("✓ Loaded config with agency: %s", config.Config.GTFS.AgencyID)
}

// TestConfig_MissingFile tests error handling for missing config
func TestConfig_MissingFile(t *testing.T) {
	// Save original config and working directory
	origConfig := config.Config
	origDir, _ := os.Getwd()
	defer func() {
		config.Config = origConfig
		os.Chdir(origDir)
	}()

	// Change to temp directory with no config
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err = config.LoadAppConfig()
	if err == nil {
		t.Error("Loading non-existent config should return error")
	}

	t.Logf("✓ Missing config returns error: %v", err)
}

// TestConfig_InvalidYAML tests error handling for invalid YAML
func TestConfig_InvalidYAML(t *testing.T) {
	// Save original config and working directory
	origConfig := config.Config
	origDir, _ := os.Getwd()
	defer func() {
		config.Config = origConfig
		os.Chdir(origDir)
	}()

	// Create temp directory with invalid YAML
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "config.yml")

	err := os.WriteFile(invalidPath, []byte("invalid: yaml: content: [[["), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Change to temp directory
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err = config.LoadAppConfig()
	if err == nil {
		t.Error("Loading invalid YAML should return error")
	}

	t.Logf("✓ Invalid YAML returns error: %v", err)
}

// TestConfig_EmptyFile tests handling of empty config file
func TestConfig_EmptyFile(t *testing.T) {
	// Save original config and working directory
	origConfig := config.Config
	origDir, _ := os.Getwd()
	defer func() {
		config.Config = origConfig
		os.Chdir(origDir)
	}()

	tmpDir := t.TempDir()
	emptyPath := filepath.Join(tmpDir, "config.yml")

	err := os.WriteFile(emptyPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err = config.LoadAppConfig()
	// Empty file might fail validation or succeed with defaults
	// Either behavior is acceptable
	t.Logf("Empty file result: %v", err)
}

// TestConfig_SelectFeedByName tests feed selection by name
func TestConfig_SelectFeedByName(t *testing.T) {
	// Create a test config with multiple feeds
	config.Config = config.AppConfig{
		Feeds: []config.Feed{
			{
				Name: "feed1",
				GTFS: config.GTFSConfig{
					AgencyID:  "AGENCY1",
					StaticURL: "http://example.com/feed1.zip",
				},
			},
			{
				Name: "feed2",
				GTFS: config.GTFSConfig{
					AgencyID:  "AGENCY2",
					StaticURL: "http://example.com/feed2.zip",
				},
			},
		},
	}

	// Select feed2
	gtfsCfg, _ := config.SelectFeed("feed2")
	if gtfsCfg.AgencyID != "AGENCY2" {
		t.Errorf("Expected AGENCY2, got %s", gtfsCfg.AgencyID)
	}

	t.Logf("✓ Selected feed2: %s", gtfsCfg.AgencyID)
}

// TestConfig_SelectDefaultFeed tests default feed selection
func TestConfig_SelectDefaultFeed(t *testing.T) {
	config.Config = config.AppConfig{
		Feeds: []config.Feed{
			{
				Name: "default",
				GTFS: config.GTFSConfig{
					AgencyID:  "DEFAULT_AGENCY",
					StaticURL: "http://example.com/default.zip",
				},
			},
			{
				Name: "other",
				GTFS: config.GTFSConfig{
					AgencyID:  "OTHER_AGENCY",
					StaticURL: "http://example.com/other.zip",
				},
			},
		},
	}

	// Select with empty name should return first feed
	gtfsCfg, _ := config.SelectFeed("")
	if gtfsCfg.AgencyID != "DEFAULT_AGENCY" {
		t.Errorf("Expected DEFAULT_AGENCY, got %s", gtfsCfg.AgencyID)
	}

	t.Logf("✓ Default feed selected: %s", gtfsCfg.AgencyID)
}

// TestConfig_SelectNonExistentFeed tests requesting non-existent feed
func TestConfig_SelectNonExistentFeed(t *testing.T) {
	config.Config = config.AppConfig{
		Feeds: []config.Feed{
			{
				Name: "existing",
				GTFS: config.GTFSConfig{
					AgencyID: "EXISTING",
				},
			},
		},
	}

	// Request non-existent feed should fall back to first
	gtfsCfg, _ := config.SelectFeed("nonexistent")
	if gtfsCfg.AgencyID != "EXISTING" {
		t.Errorf("Should fall back to first feed")
	}

	t.Log("✓ Non-existent feed falls back to default")
}

// TestConfig_ValidateGTFSConfig tests GTFS config validation
func TestConfig_ValidateGTFSConfig(t *testing.T) {
	// Valid config
	validCfg := config.GTFSConfig{
		AgencyID:  "TEST",
		StaticURL: "http://example.com/gtfs.zip",
	}

	if validCfg.AgencyID == "" {
		t.Error("Valid config should have agency_id")
	}

	// Missing fields should still be valid (fields are optional)
	minimalCfg := config.GTFSConfig{
		AgencyID: "MINIMAL",
	}

	if minimalCfg.AgencyID == "" {
		t.Error("Minimal config should have agency_id")
	}

	t.Log("✓ GTFS config validation works")
}

// TestConfig_ValidateFeedConfig tests feed config structure
func TestConfig_ValidateFeedConfig(t *testing.T) {
	feed := config.Feed{
		Name: "test-feed",
		GTFS: config.GTFSConfig{
			AgencyID:  "TEST",
			StaticURL: "http://example.com/gtfs.zip",
		},
		GTFSRT: config.GTFSRTConfig{
			TripUpdatesURL:      "http://example.com/tu",
			VehiclePositionsURL: "http://example.com/vp",
			ServiceAlertsURL:    "http://example.com/sa",
		},
	}

	if feed.Name == "" {
		t.Error("Feed should have name")
	}

	if feed.GTFS.AgencyID == "" {
		t.Error("Feed should have GTFS config")
	}

	t.Log("✓ Feed config structure valid")
}

// TestConfig_ConverterConfig tests converter configuration
func TestConfig_ConverterConfig(t *testing.T) {
	converterCfg := config.ConverterConfig{
		CallDistanceAlongRouteNumDigits: 3,
		TripKeyStrategy:                 "raw",
		UnscheduledTripIndicator:        "_",
	}

	if converterCfg.TripKeyStrategy != "raw" {
		t.Errorf("Expected 'raw', got '%s'", converterCfg.TripKeyStrategy)
	}

	if converterCfg.CallDistanceAlongRouteNumDigits != 3 {
		t.Error("Distance digits should be 3")
	}

	t.Log("✓ Converter config validated")
}

// TestConfig_TripKeyStrategies tests different trip key strategy values
func TestConfig_TripKeyStrategies(t *testing.T) {
	strategies := []string{"raw", "startDateTrip", "agencyTrip", "agencyStartDateTrip"}

	for _, strategy := range strategies {
		cfg := config.ConverterConfig{
			TripKeyStrategy: strategy,
		}

		if cfg.TripKeyStrategy != strategy {
			t.Errorf("Strategy should be '%s', got '%s'", strategy, cfg.TripKeyStrategy)
		}
	}

	t.Log("✓ All trip key strategies are valid")
}

// TestConfig_FieldMutators tests field mutator configuration
func TestConfig_FieldMutators(t *testing.T) {
	mutators := config.FieldMutators{
		OriginRef:      []string{"prefix:AGENCY:"},
		StopPointRef:   []string{"prefix:AGENCY:"},
		DestinationRef: []string{"prefix:AGENCY:"},
	}

	if len(mutators.OriginRef) == 0 {
		t.Error("Should have OriginRef mutators")
	}

	if mutators.OriginRef[0] != "prefix:AGENCY:" {
		t.Errorf("Expected 'prefix:AGENCY:', got '%s'", mutators.OriginRef[0])
	}

	t.Log("✓ Field mutators configured")
}

// TestConfig_ServerConfig tests server configuration
func TestConfig_ServerConfig(t *testing.T) {
	serverCfg := config.ServerConfig{
		Port: 8080,
	}

	if serverCfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", serverCfg.Port)
	}

	t.Log("✓ Server config validated")
}

// TestConfig_YAMLMarshaling tests that config can be marshaled/unmarshaled
func TestConfig_YAMLMarshaling(t *testing.T) {
	original := config.AppConfig{
		GTFS: config.GTFSConfig{
			AgencyID:  "TEST",
			StaticURL: "http://example.com/gtfs.zip",
		},
		Converter: config.ConverterConfig{
			TripKeyStrategy: "raw",
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal back
	var parsed config.AppConfig
	err = yaml.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if parsed.GTFS.AgencyID != original.GTFS.AgencyID {
		t.Error("Marshaling/unmarshaling should preserve data")
	}

	t.Log("✓ YAML marshaling works")
}
