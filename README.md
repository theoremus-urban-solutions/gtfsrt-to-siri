# GTFS-Realtime to SIRI Converter

A lightweight library for transforming GTFS-Realtime transit data into SIRI (Standard Interface for Realtime Information) format. The implementation follows Entur's latest public specification.

**Features:**
- **On-demand conversion** – No caching layer; designed for integration into larger systems
- **Fast HTTP calls** – 10-second timeouts with 3 retries
- **Standardized SIRI output** – Includes `ProducerRef` (codespace) in all responses
- **Modular architecture** – Clean separation: `siri/`, `gtfs/`, `gtfsrt/`, `converter/`, `formatter/`, `tracking/`, `utils/`

## Acknowledgments

This implementation follows [Entur's Nordic SIRI Profile](https://enturas.atlassian.net/wiki/spaces/PUBLIC/pages/637370373/) specification. Initial project structure was inspired by MTA's [GTFS-Realtime to SIRI Converter](https://github.com/availabs/MTA_Subway_GTFS-Realtime_to_SIRI_Converter).

## Installation

```bash
go build -o gtfsrt-to-siri ./cmd/gtfsrt-to-siri/
```

## CLI Usage

### Basic Commands

**Vehicle Monitoring (VM)**
```bash
./gtfsrt-to-siri -mode=oneshot -call=vm -format=xml -modules=tu,vp
./gtfsrt-to-siri -call=vm -format=json  # JSON output
```

**Estimated Timetable (ET)**
```bash
./gtfsrt-to-siri -call=et -format=xml -modules=tu,vp
./gtfsrt-to-siri -call=et -monitoringRef=STOP_ID -lineRef=ROUTE_ID -directionRef=0
```

**Situation Exchange (SX)**
```bash
./gtfsrt-to-siri -call=sx -format=xml -modules=alerts
```

### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-mode` | Execution mode | `oneshot` |
| `-call` | SIRI module: `vm`, `et`, `sx` | `vm` |
| `-format` | Output format: `json`, `xml` | `json` |
| `-modules` | GTFS-RT modules to fetch: `tu`, `vp`, `alerts` | `tu,vp` |
| `-feed` | Feed name from `config.yml` feeds list | (first feed) |
| `-tripUpdates` | Override TripUpdates URL | (from config) |
| `-vehiclePositions` | Override VehiclePositions URL | (from config) |
| `-serviceAlerts` | Override ServiceAlerts URL | (from config) |
| `-monitoringRef` | Stop ID filter (optional for ET) | |
| `-lineRef` | Filter by route/line | |
| `-directionRef` | Filter by direction: `0` or `1` | |

## Library Usage

This library is **data-source agnostic** and designed for integration into servers (Kafka-based, HTTP APIs, etc.). You provide raw GTFS and GTFS-RT data, the library handles conversion.

### Architecture Overview

```
Your Server
    ↓ (fetch from MinIO/HTTP/files)
GTFS zip bytes + GTFS-RT protobuf bytes
    ↓ (parse once)
GTFSIndex (cached in memory) + GTFSRTWrapper
    ↓ (convert)
Converter → SIRI Response (VM/ET/SX)
```

### Key Principle: Cache the GTFS Index

**⚠️ IMPORTANT:** Parse GTFS once at startup and keep `*gtfs.GTFSIndex` in memory. GTFS is static data - parsing the zip on every request is wasteful (500ms-2s vs <1ms).

### Server Integration Example

```go
import (
    "github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
    "github.com/theoremus-urban-solutions/gtfsrt-to-siri/formatter"
    "github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfs"
    "github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
)

type Server struct {
    gtfsIndex *gtfs.GTFSIndex  // ✅ Parse once, reuse for all requests
    opts      converter.ConverterOptions
}

// Initialize server - parse GTFS once
func (s *Server) Init() error {
    // 1. Fetch GTFS from your source (MinIO, HTTP, file, etc.)
    gtfsZipBytes, err := fetchFromMinIO("path/to/gtfs.zip")
    if err != nil {
        return err
    }

    // 2. Parse GTFS once into memory (expensive operation)
    s.gtfsIndex, err = gtfs.NewGTFSIndexFromBytes(gtfsZipBytes, "YOUR_AGENCY_ID")
    if err != nil {
        return err
    }

    // 3. Configure converter options
    s.opts = converter.ConverterOptions{
        AgencyID:       "YOUR_AGENCY_ID",
        ReadIntervalMS: 30000,
        FieldMutators: converter.FieldMutators{
            StopPointRef:   []string{"OLD", "NEW"},  // optional
            OriginRef:      []string{"OLD", "NEW"},  // optional
            DestinationRef: []string{"OLD", "NEW"},  // optional
        },
    }

    return nil
}

// Handle each Kafka message / HTTP request
func (s *Server) ProcessRealtimeData(protobufBytes []byte) ([]byte, error) {
    // 1. Parse GTFS-RT protobuf (fast - only current data)
    rt, err := gtfsrt.NewGTFSRTWrapper(
        protobufBytes,  // TripUpdates
        protobufBytes,  // VehiclePositions (or separate bytes)
        nil,            // ServiceAlerts (optional)
    )
    if err != nil {
        return nil, err
    }

    // 2. Create converter (reuses cached GTFS index)
    conv := converter.NewConverter(s.gtfsIndex, rt, s.opts)

    // 3. Generate SIRI response
    response := conv.GetCompleteVehicleMonitoringResponse()
    // or: et := conv.BuildEstimatedTimetable()
    // or: sx := conv.BuildSituationExchange()

    // 4. Format as XML or JSON
    rb := formatter.NewResponseBuilder()
    return rb.BuildJSON(response), nil  // or rb.BuildXML(response)
}

// Update GTFS when it changes (daily/weekly)
func (s *Server) UpdateGTFS() error {
    gtfsZipBytes, err := fetchFromMinIO("path/to/gtfs.zip")
    if err != nil {
        return err
    }

    newIndex, err := gtfs.NewGTFSIndexFromBytes(gtfsZipBytes, s.opts.AgencyID)
    if err != nil {
        return err
    }

    s.gtfsIndex = newIndex  // Atomic swap (consider mutex for concurrent access)
    return nil
}
```

### Quick Start: One-Shot Conversion

For simple scripts or testing:

```go
// Fetch GTFS
gtfsZipBytes, _ := os.ReadFile("gtfs.zip")
gtfsIndex, _ := gtfs.NewGTFSIndexFromBytes(gtfsZipBytes, "AGENCY_ID")

// Fetch GTFS-RT (implement your own fetching logic)
tuBytes, _ := fetchFromHTTP("http://example.com/tripupdates")
vpBytes, _ := fetchFromHTTP("http://example.com/vehiclepositions")
saBytes, _ := fetchFromHTTP("http://example.com/alerts")

// Create wrapper
rt, _ := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, saBytes)

// Convert
opts := converter.ConverterOptions{AgencyID: "AGENCY_ID", ReadIntervalMS: 30000}
conv := converter.NewConverter(gtfsIndex, rt, opts)
response := conv.GetCompleteVehicleMonitoringResponse()

// Format
rb := formatter.NewResponseBuilder()
xmlBytes := rb.BuildXML(response)
```

### API Reference

**Vehicle Monitoring**
```go
response := conv.GetCompleteVehicleMonitoringResponse()
// Returns *siri.VehicleMonitoringResponse with all vehicles
```

**Estimated Timetable**
```go
et := conv.BuildEstimatedTimetable()
// Filter if needed:
filtered := formatter.FilterEstimatedTimetable(et, stopID, lineRef, directionRef)
// Wrap in response:
response := formatter.WrapEstimatedTimetableResponse(filtered, agencyID)
```

**Situation Exchange**
```go
sx := conv.BuildSituationExchange()
timestamp := rt.GetTimestampForFeedMessage()
response := formatter.WrapSituationExchangeResponse(sx, timestamp, agencyID)
```

### Performance Notes

- **GTFS parsing**: 500ms-2s (do once at startup)
- **GTFS-RT parsing**: 10-50ms per message
- **Conversion**: <1ms with cached GTFS index
- **Formatting**: 5-20ms for XML/JSON

**Memory footprint:** ~100-200MB for typical GTFS index (10-20K stops, 1K routes)

## Testing

Run the test suite:
```bash
# All tests
go test ./tests/...

# Unit tests only
go test ./tests/unit -v

# Integration tests only
go test ./tests/integration -v
```

The test suite includes:
- **68 tests** (25 integration + 43 unit) covering all critical packages
- **67.1% code coverage** (converter: 76%, formatter: ~75%, gtfsrt: 72%, gtfs: 66%, utils: 80%)
- **Real-world data** from Sofia Public Transport (1,237 vehicles)
- **Critical regression tests** for VehicleMode, delay calculation, and start_date handling
- **GitHub Actions CI/CD** with automated testing and coverage tracking

See [docs/TESTING_IMPLEMENTATION_COMPLETE.md](docs/TESTING_IMPLEMENTATION_COMPLETE.md) and [docs/CODE_COVERAGE.md](docs/CODE_COVERAGE.md) for detailed reports.

## Configuration

Edit `config.yml`:

```yaml
gtfs:
  agency_id: "AGENCY"
  url: "http://example.com/gtfs.zip"
  
feeds:
  - name: "main"
    gtfsrt:
      trip_updates: "http://example.com/tripupdates"
      vehicle_positions: "http://example.com/vehiclepositions"
      service_alerts: "http://example.com/alerts"

converter:
  field_mutators:
    stop_point_ref:
      - type: "prefix"
        value: "AGENCY:"
```

## SIRI Modules

- **VM (Vehicle Monitoring)**: Real-time vehicle positions and trip progress
- **ET (Estimated Timetable)**: Stop-level arrival/departure predictions for routes
- **SX (Situation Exchange)**: Service alerts and disruptions

## References

- [Nordic SIRI Profile](https://enturas.atlassian.net/wiki/spaces/PUBLIC/pages/637370373/)
- [GTFS-Realtime Specification](https://gtfs.org/realtime/)

