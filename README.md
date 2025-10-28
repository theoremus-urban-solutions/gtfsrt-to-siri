# GTFS-Realtime to SIRI Converter

A lightweight library for transforming GTFS-Realtime transit data into SIRI (Standard Interface for Realtime Information) format. The implementation follows Entur's latest public specification.

Sister project: [SIRI to GTFS-Realtime](https://github.com/theoremus-urban-solutions/siri-to-gtfsrt)

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

**⚠️ CRITICAL FOR PERFORMANCE:** Always cache GTFS static data! Parsing the GTFS zip on every request is extremely wasteful.

**Performance Impact:**
- Without caching: 1,200-1,800ms per request (fetch + parse GTFS static every time)
- With caching: 100-150ms per request (**10x faster, 89% time reduction**)

The library provides two approaches for GTFS static caching:

#### Option 1: In-Memory Caching (Simple)

Parse GTFS once at startup and keep `*gtfs.GTFSIndex` in memory. Best for single-instance deployments.

```go
import (
    "github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
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

    // 2. Parse GTFS once into memory (takes 300-400ms)
    s.gtfsIndex, err = gtfs.NewGTFSIndexFromBytes(gtfsZipBytes, "YOUR_AGENCY_ID")
    if err != nil {
        return err
    }

    // 3. Configure converter options
    s.opts = converter.ConverterOptions{
        AgencyID:       "YOUR_AGENCY_ID",
        ReadIntervalMS: 30000,
    }

    return nil
}

// Handle each Kafka message / HTTP request
func (s *Server) ProcessRealtimeData(tuBytes, vpBytes []byte) ([]byte, error) {
    // 1. Parse GTFS-RT protobuf (fast - 10-20ms)
    rt, err := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, nil)
    if err != nil {
        return nil, err
    }

    // 2. Create converter using cached GTFS index (explicit API)
    conv := converter.NewConverterWithCachedGTFS(s.gtfsIndex, rt, s.opts)

    // 3. Generate SIRI response (fast - <100ms with cached index)
    response := conv.GetCompleteVehicleMonitoringResponse()

    // 4. Format as JSON
    rb := formatter.NewResponseBuilder()
    return rb.BuildJSON(response), nil
}
```

#### Option 2: Disk-Based Caching (Production)

For production deployments, persist the parsed GTFS index to disk. This allows:
- Faster startup (load from cache instead of re-parsing)
- Shared cache across multiple instances
- Survive restarts without re-downloading/re-parsing

```go
import (
    "os"
    "sync"
    "time"

    "github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfs"
    "github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
    "github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
)

type Server struct {
    mu        sync.RWMutex
    gtfsIndex *gtfs.GTFSIndex
    opts      converter.ConverterOptions
}

// Initialize with disk-based cache
func (s *Server) Init(staticURL, cacheFile string) error {
    // Try to load from cache first
    if index, err := gtfs.DeserializeIndexFromFile(cacheFile); err == nil {
        log.Println("Loaded GTFS index from cache")
        s.gtfsIndex = index

        // Start background refresh (e.g., daily)
        go s.refreshGTFSPeriodically(staticURL, cacheFile, 24*time.Hour)
        return nil
    }

    // Cache miss - fetch and parse fresh data
    log.Println("Cache miss, fetching fresh GTFS data...")
    return s.updateGTFSCache(staticURL, cacheFile)
}

func (s *Server) updateGTFSCache(staticURL, cacheFile string) error {
    // 1. Fetch GTFS static data
    gtfsZipBytes, err := gtfs.FetchGTFSData(staticURL)
    if err != nil {
        return err
    }

    // 2. Parse into memory
    index, err := gtfs.NewGTFSIndexFromBytes(gtfsZipBytes, "AGENCY_ID")
    if err != nil {
        return err
    }

    // 3. Save to disk cache
    if err := gtfs.SerializeIndexToFile(index, cacheFile); err != nil {
        log.Printf("Warning: failed to save cache: %v", err)
    }

    // 4. Update in-memory cache (thread-safe)
    s.mu.Lock()
    s.gtfsIndex = index
    s.mu.Unlock()

    log.Printf("GTFS cache updated: %d trips, %d stops",
        len(index.TripStopSeq), len(index.StopNames))
    return nil
}

func (s *Server) refreshGTFSPeriodically(staticURL, cacheFile string, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for range ticker.C {
        if err := s.updateGTFSCache(staticURL, cacheFile); err != nil {
            log.Printf("Failed to refresh GTFS cache: %v", err)
        }
    }
}

// Handle each request (thread-safe)
func (s *Server) ProcessRealtimeData(tuBytes, vpBytes []byte) ([]byte, error) {
    // 1. Parse GTFS-RT (fast)
    rt, err := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, nil)
    if err != nil {
        return nil, err
    }

    // 2. Get cached GTFS index (thread-safe read)
    s.mu.RLock()
    cachedIndex := s.gtfsIndex
    s.mu.RUnlock()

    // 3. Convert using cached index
    conv := converter.NewConverterWithCachedGTFS(cachedIndex, rt, s.opts)
    response := conv.GetCompleteVehicleMonitoringResponse()

    // 4. Format
    rb := formatter.NewResponseBuilder()
    return rb.BuildJSON(response), nil
}
```

**Cache serialization API:**
```go
// Save to file
err := gtfs.SerializeIndexToFile(index, "/path/to/cache.gob")

// Load from file
index, err := gtfs.DeserializeIndexFromFile("/path/to/cache.gob")

// Save to custom storage (S3, MinIO, etc.)
var buf bytes.Buffer
err := gtfs.SerializeIndexToWriter(index, &buf)
// Upload buf.Bytes() to your storage

// Load from custom storage
index, err := gtfs.DeserializeIndexFromReader(reader)
```

**See [CACHING.md](CACHING.md) for detailed examples including:**
- HTTP conditional GET (ETag-based cache validation)
- Multi-agency deployments
- Cache invalidation strategies
- Complete production-ready implementation

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

**With GTFS Static Caching (Recommended):**
- **First load**: 1,300-1,800ms (fetch + parse GTFS static, one-time cost)
- **Subsequent requests**: 100-150ms total
  - GTFS-RT fetch: 30-70ms
  - GTFS-RT parse: 15-30ms
  - Conversion: 60-120ms
  - Formatting: 15-20ms
- **Speedup: 10x faster** (89% time reduction)

**Without Caching (Not Recommended):**
- Every request: 1,800-2,600ms (includes 1,200-1,800ms GTFS overhead)

**Memory footprint:**
- Typical GTFS index: 30-50MB (26K trips, 4K stops, 179 routes)
- Large agencies: 100-200MB (100K+ trips)
- Cache file size: ~32MB (compressed with gob encoding)

**Measured with real data** (Sofia Public Transport GTFS):
- GTFS static fetch: 900-1,400ms
- GTFS static parse: 300-420ms
- Cache serialize: 230ms
- Cache deserialize: 130ms (9.9x faster than parse)
- ET conversion (1,255 trips): 67-114ms (0.05-0.09ms per trip)

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

