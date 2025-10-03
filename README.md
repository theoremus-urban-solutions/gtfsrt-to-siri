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

### Setup

```go
import (
    gtfsrtsiri "mta/gtfsrt-to-siri"
)

// Load configuration
gtfsrtsiri.InitLogging()
if err := gtfsrtsiri.LoadAppConfig(); err != nil {
    panic(err)
}

// Select feed (or use default)
gtfsCfg, rtCfg := gtfsrtsiri.SelectFeed("")

// Initialize GTFS and GTFS-RT
gtfs, _ := gtfsrtsiri.NewGTFSIndexFromConfig(gtfsCfg)
rt := gtfsrtsiri.NewGTFSRTWrapper(
    rtCfg.TripUpdatesURL,
    rtCfg.VehiclePositionsURL,
    rtCfg.ServiceAlertsURL,
)

// Fetch latest data
rt.Refresh()

// Create converter
conv := gtfsrtsiri.NewConverter(gtfs, rt, gtfsrtsiri.Config)
cache := gtfsrtsiri.NewConverterCache(conv)
```

### Generate Responses

**Vehicle Monitoring**
```go
buf, err := cache.GetVehicleMonitoringResponse(map[string]string{}, "xml")
// Returns SIRI VehicleMonitoring XML/JSON
```

**Estimated Timetable**
```go
params := map[string]string{
    "monitoringref": "STOP_ID",  // optional
    "lineref": "ROUTE_ID",       // optional
    "directionref": "0",         // optional
}
buf, err := cache.GetEstimatedTimetableResponse(params, "json")
```

**Situation Exchange**
```go
buf, err := cache.GetSituationExchangeResponse("xml")
// Returns SIRI SituationExchange with alerts
```

### Core Types

**Converter**
```go
type Converter struct {
    GTFS   *GTFSIndex
    GTFSRT *GTFSRTWrapper
    Cfg    *AppConfig
}

// Build responses
func (c *Converter) BuildVehicleMonitoring() []VehicleActivity
func (c *Converter) BuildEstimatedTimetable() EstimatedTimetable
func (c *Converter) BuildSituationExchange() []PtSituationElement
```

**ConverterCache**
```go
type ConverterCache struct {
    converter *Converter
    rb        *responseBuilder
}

// Generate formatted responses (XML/JSON)
func (cc *ConverterCache) GetVehicleMonitoringResponse(params map[string]string, format string) ([]byte, error)
func (cc *ConverterCache) GetEstimatedTimetableResponse(params map[string]string, format string) ([]byte, error)
func (cc *ConverterCache) GetSituationExchangeResponse(format string) ([]byte, error)
```

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

