# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library and CLI tool that converts GTFS-Realtime (General Transit Feed Specification - Realtime) data to SIRI (Standard Interface for Realtime Information) format, following Entur's Nordic SIRI Profile specification.

**Key principle:** The library is **data-source agnostic**. It accepts raw bytes (GTFS zip, GTFS-RT protobuf) and handles conversion. Data fetching (HTTP, Kafka, MinIO, etc.) is the responsibility of integrating applications.

## Development Commands

### Building
```bash
make build              # Build for current platform → bin/gtfsrt-to-siri
make build-all          # Build for Linux, macOS (Intel/ARM), Windows
go build -o gtfsrt-to-siri ./cmd/gtfsrt-to-siri/  # Direct build
```

### Testing
```bash
make test               # Run all tests (integration + unit)
make test-unit          # Unit tests only
make test-integration   # Integration tests only
make test-race          # Run tests with race detector
make coverage           # Generate coverage report → coverage.html
go test ./tests/...     # Direct test execution
go test ./tests/integration/converter_vm_test.go -v  # Run single test file
go test -run TestConverterVM ./tests/integration/    # Run specific test
```

### Code Quality
```bash
make lint               # Run golangci-lint
make fmt                # Format code with go fmt
make vet                # Run go vet
make todo               # Show TODO/FIXME comments in code
make benchmark          # Run performance benchmarks
```

### Running the CLI
```bash
# Show help
./gtfsrt-to-siri -help

# From URLs: Vehicle Monitoring + Estimated Timetable (XML - default)
./gtfsrt-to-siri \
  -tripUpdates=https://example.com/tu \
  -vehiclePositions=https://example.com/vp

# From URLs: Estimated Timetable only (JSON)
./gtfsrt-to-siri -tripUpdates=https://example.com/tu -format=json

# From URLs: All formats (VM + ET + SX)
./gtfsrt-to-siri \
  -tripUpdates=https://example.com/tu \
  -vehiclePositions=https://example.com/vp \
  -alerts=https://example.com/alerts

# From config file
./gtfsrt-to-siri -read-from-config
```

**CLI Design:**
- **URL-first approach**: Provide feed URLs directly via flags OR use `-read-from-config`
- Must provide **at least one URL** (tripUpdates, vehiclePositions, or alerts) OR `-read-from-config`
- URL flags and `-read-from-config` are **mutually exclusive**
- `-format` defaults to `xml` (can be `json` or `xml`)
- The CLI automatically generates **all applicable SIRI formats** based on which URLs are provided:
  - `-tripUpdates` only → Estimated Timetable (ET)
  - `-alerts` only → Situation Exchange (SX)
  - `-tripUpdates` + `-vehiclePositions` → Vehicle Monitoring (VM) + Estimated Timetable (ET)
  - All three URLs → VM + ET + SX (all three formats)
- GTFS static data is always fetched from `config.yml` (gtfs.staticURL)
- No filtering support - outputs complete unfiltered data

## Architecture

### Core Package Structure

```
gtfs/           → GTFS static data parsing & indexing (routes, stops, trips, shapes)
gtfsrt/         → GTFS-Realtime protobuf parsing (trip updates, vehicle positions, alerts)
converter/      → Core conversion logic (GTFS-RT → SIRI)
siri/           → SIRI data structures (VM, ET, SX types)
formatter/      → SIRI output formatting (JSON, XML)
tracking/       → Vehicle position tracking and estimation
utils/          → Utilities (time, distance, constants)
config/         → Configuration loading (config.yml)
cmd/            → CLI application
tests/          → Test suites (integration + unit)
```

### Data Flow

```
1. GTFS zip bytes → gtfs.NewGTFSIndexFromBytes() → *gtfs.GTFSIndex (cache this!)
2. GTFS-RT protobuf bytes → gtfsrt.NewGTFSRTWrapper() → *gtfsrt.GTFSRTWrapper
3. Index + Wrapper + Options → converter.NewConverter() → *converter.Converter
4. Converter methods:
   - GetCompleteVehicleMonitoringResponse() → VM response
   - BuildEstimatedTimetable() → ET data
   - BuildSituationExchange() → SX data
5. formatter.ResponseBuilder → BuildJSON() / BuildXML()
```

### Critical Performance Pattern

**⚠️ ALWAYS cache the GTFS index in memory.** GTFS is static data (routes, stops, schedules) that changes infrequently (daily/weekly). Parse it **once at startup** and reuse for all requests.

- GTFS parsing: 500ms-2s (expensive)
- GTFS-RT parsing: 10-50ms (per message)
- Conversion with cached index: <1ms (fast!)

```go
// ✅ CORRECT - Parse once, reuse for all requests
type Server struct {
    gtfsIndex *gtfs.GTFSIndex  // Cached at startup
}

func (s *Server) Init() {
    gtfsBytes := fetchFromSource()
    s.gtfsIndex, _ = gtfs.NewGTFSIndexFromBytes(gtfsBytes, "AGENCY_ID")
}

func (s *Server) HandleRequest(rtBytes []byte) {
    rt, _ := gtfsrt.NewGTFSRTWrapper(rtBytes, nil, nil)
    conv := converter.NewConverter(s.gtfsIndex, rt, opts)  // Reuses cached index
    return conv.GetCompleteVehicleMonitoringResponse()
}
```

```go
// ❌ WRONG - Parsing GTFS on every request (500ms-2s overhead!)
func HandleRequest(gtfsBytes, rtBytes []byte) {
    gtfsIndex, _ := gtfs.NewGTFSIndexFromBytes(gtfsBytes, "AGENCY")  // DON'T DO THIS
    rt, _ := gtfsrt.NewGTFSRTWrapper(rtBytes, nil, nil)
    // ...
}
```

### SIRI Modules

The converter supports three SIRI modules:

1. **VM (Vehicle Monitoring)**: Real-time vehicle positions, journey progress, next stops
   - Requires: GTFS + TripUpdates + VehiclePositions
   - Use: `conv.GetCompleteVehicleMonitoringResponse()`

2. **ET (Estimated Timetable)**: Stop-level arrival/departure predictions
   - Requires: GTFS + TripUpdates
   - Use: `conv.BuildEstimatedTimetable()` + filtering + wrapping

3. **SX (Situation Exchange)**: Service alerts and disruptions
   - Requires: GTFS + ServiceAlerts
   - Use: `conv.BuildSituationExchange()` + wrapping

### Configuration System

The library itself is config-free and accepts options via structs. The CLI application uses `config.yml` for convenience:

```yaml
gtfs:
  agency_id: SOFIA           # Used in SIRI reference formatting
  staticURL: http://...      # Where to fetch GTFS zip

gtfsrt:
  tripUpdatesURL: http://...
  vehiclePositionsURL: http://...
  serviceAlertsURL: http://...
  readIntervalMS: 30000      # For ValidUntil calculation

converter:
  field_mutators:            # Optional string replacements in SIRI fields
    stop_point_ref:
      - type: "prefix"
        value: "AGENCY:"
```

### Field Mutators

Field mutators allow string replacement in SIRI reference fields (useful for ID transformations):

```go
opts := converter.ConverterOptions{
    FieldMutators: converter.FieldMutators{
        StopPointRef:   []string{"OLD_STOP_1", "NEW_STOP_1", "OLD_STOP_2", "NEW_STOP_2"},
        OriginRef:      []string{"OLD_ORIGIN", "NEW_ORIGIN"},
        DestinationRef: []string{"OLD_DEST", "NEW_DEST"},
    },
}
```

Format: `[from1, to1, from2, to2, ...]` - pairs of old/new values for replacement.

### Thread Safety

- **Converter instances**: NOT thread-safe. Create a new converter per request.
- **GTFS index and GTFS-RT wrapper**: Thread-safe for reads. Can be safely shared across goroutines.

### Testing

The project has comprehensive test coverage:
- **68 tests total** (25 integration + 43 unit tests)
- **67.1% overall coverage** (converter: 76%, formatter: ~75%, gtfsrt: 72%, gtfs: 66%, utils: 80%)
- Real-world test data from Sofia Public Transport (1,237 vehicles)
- Critical regression tests for VehicleMode, delay calculation, start_date handling

When adding features:
- Add unit tests in `tests/unit/<package>/` for isolated logic
- Add integration tests in `tests/integration/` for end-to-end scenarios
- Use existing test data in `testdata/` or add new fixtures as needed

### Common Patterns

**Creating a complete conversion pipeline:**
```go
// 1. Load GTFS (once)
gtfsBytes := fetchGTFS()
gtfsIndex, _ := gtfs.NewGTFSIndexFromBytes(gtfsBytes, "AGENCY")

// 2. Parse GTFS-RT (per request)
tuBytes, vpBytes, saBytes := fetchRealtime()
rt, _ := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, saBytes)

// 3. Configure converter
opts := converter.ConverterOptions{
    AgencyID:       "AGENCY",
    ReadIntervalMS: 30000,
}

// 4. Convert
conv := converter.NewConverter(gtfsIndex, rt, opts)
vmResponse := conv.GetCompleteVehicleMonitoringResponse()

// 5. Format output
rb := formatter.NewResponseBuilder()
xmlBytes := rb.BuildXML(vmResponse)
jsonBytes := rb.BuildJSON(vmResponse)
```

**Filtering Estimated Timetable:**
```go
et := conv.BuildEstimatedTimetable()
// Filter parameters are case-insensitive and support partial matches
filtered := formatter.FilterEstimatedTimetable(et, stopID, lineRef, directionRef)
response := formatter.WrapEstimatedTimetableResponse(filtered, agencyID)
```

**Using Wrapping Functions:**
```go
// Wrap ET response
etResponse := formatter.WrapEstimatedTimetableResponse(et, agencyID)

// Wrap SX response
timestamp := rt.GetTimestampForFeedMessage()
sxResponse := formatter.WrapSituationExchangeResponse(sx, timestamp, agencyID)
```

### Warning System

The converter includes a **WarningAggregator** that consolidates data quality issues and logs them in a user-friendly format:

```go
// Warnings are collected during conversion and logged at the end
conv := converter.NewConverter(gtfsIndex, rt, opts)
response := conv.GetCompleteVehicleMonitoringResponse()
// Warnings are automatically logged via conv.Warnings.LogAll()
```

Common warning types:
- **VM warnings**: `no_route_id`, `no_lat_lon`, `no_onward_stops`, `origin_stop_no_name`
- **ET warnings**: `no_start_date`, `trip_not_in_static`, `no_static_times`, `no_stop_time_updates`
- **SX warnings**: `no_summary`, `no_description`, `stop_not_found`

Each warning includes:
- Count of occurrences
- Up to 3 example IDs
- Description of the issue
- Action taken (fallback behavior)

Example log output:
```
Feed TU for agency SOFIA has missing route_id fields (12 occurrences).
Building SIRI output with fallback 'UNKNOWN'. Examples: trip_123, trip_456, trip_789
```

### Key Files to Know

- `converter/converter.go`: Main converter struct and constructor
- `converter/vm.go`: Vehicle Monitoring conversion logic
- `converter/et.go`: Estimated Timetable conversion logic
- `converter/sx.go`: Situation Exchange conversion logic
- `converter/warnings.go`: Warning aggregation and logging system
- `gtfs/loader.go`: GTFS zip parsing and validation
- `gtfs/index.go`: GTFS in-memory index structure
- `gtfsrt/wrapper.go`: GTFS-RT protobuf parsing and access methods
- `formatter/xml.go`: XML serialization (complex due to SIRI schema)
- `formatter/json.go`: JSON serialization
- `formatter/wrapper.go`: Response wrapping and filtering utilities
- `cmd/gtfsrt-to-siri/main.go`: CLI entry point

### External Dependencies

- `github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs` - GTFS-RT protobuf definitions
- `google.golang.org/protobuf` - Protobuf parsing
- `gopkg.in/yaml.v3` - Config file parsing
- `github.com/go-playground/validator/v10` - Struct validation

### SIRI Reference Formatting

All SIRI references follow the pattern `{AgencyID}:{Type}:{ID}`:
- LineRef: `SOFIA:Line:84`
- StopPointRef: `SOFIA:Quay:1234`
- VehicleRef: `SOFIA:VehicleRef:5678`
- OriginRef/DestinationRef: `SOFIA:Quay:1000`

The AgencyID is critical for proper SIRI compliance and must match the GTFS agency_id.
