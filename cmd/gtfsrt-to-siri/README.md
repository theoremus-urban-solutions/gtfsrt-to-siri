# GTFS-RT to SIRI CLI

Thin CLI client that fetches GTFS and GTFS-RT data and converts to SIRI format.

## Purpose

The CLI is a convenience wrapper around the core library. It:
1. Loads `config.yml` (URLs, agency ID, mutators)
2. Fetches GTFS zip via HTTP or file
3. Fetches GTFS-RT protobuf via HTTP
4. Calls library with raw bytes
5. Formats output (JSON/XML)
6. Prints to stdout

**Note:** For production systems, integrate the library directly rather than using the CLI. See main [README.md](../../README.md) for server integration examples.

## Installation

```bash
go build -o gtfsrt-to-siri ./cmd/gtfsrt-to-siri/
```

## Usage

```bash
# Vehicle Monitoring (JSON)
./gtfsrt-to-siri -call=vm -format=json

# Estimated Timetable (XML)
./gtfsrt-to-siri -call=et -format=xml

# Situation Exchange
./gtfsrt-to-siri -call=sx

# ET with filters
./gtfsrt-to-siri -call=et -monitoringRef=STOP_123 -lineRef=ROUTE_1 -directionRef=0
```

### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-call` | SIRI call type: `vm`, `et`, `sx` | `vm` |
| `-format` | Output format: `json` or `xml` | `json` |
| `-modules` | GTFS-RT modules to fetch: `tu`, `vp`, `alerts` | `tu,vp` |
| `-feed` | Feed name from config.feeds[] | (first feed) |
| `-tripUpdates` | Override TripUpdates URL | (from config) |
| `-vehiclePositions` | Override VehiclePositions URL | (from config) |
| `-serviceAlerts` | Override ServiceAlerts URL | (from config) |
| `-monitoringRef` | Stop ID filter (ET only) | |
| `-lineRef` | Route/line filter | |
| `-directionRef` | Direction filter: `0` or `1` | |

## Configuration

Create `config.yml`:

```yaml
gtfs:
  agency_id: "AGENCY"
  url: "http://example.com/gtfs.zip"
  # OR use file path:
  # url: "./data/gtfs.zip"

feeds:
  - name: "main"
    gtfsrt:
      trip_updates: "http://example.com/tripupdates"
      vehicle_positions: "http://example.com/vehiclepositions"
      service_alerts: "http://example.com/alerts"
      timeout_seconds: 10
      read_interval_ms: 30000

converter:
  field_mutators:
    stop_point_ref:
      - "OLD_STOP_1"
      - "NEW_STOP_1"
    origin_ref:
      - "OLD_ORIGIN"
      - "NEW_ORIGIN"
    destination_ref:
      - "OLD_DEST"
      - "NEW_DEST"
```

## Examples

### Basic VM Call

```bash
./gtfsrt-to-siri -call=vm -format=json > output.json
```

### ET with Stop Filter

```bash
./gtfsrt-to-siri -call=et -monitoringRef=STOP_123 -format=xml
```

### URL Overrides

```bash
./gtfsrt-to-siri \
  -tripUpdates="http://custom-url.com/tu" \
  -vehiclePositions="http://custom-url.com/vp" \
  -call=vm
```

### Select Specific Modules

```bash
# Only fetch vehicle positions (no trip updates)
./gtfsrt-to-siri -modules=vp -call=vm

# Only fetch alerts
./gtfsrt-to-siri -modules=alerts -call=sx
```

## Architecture

```
┌──────────────┐
│   config.yml │
└──────┬───────┘
       │
       ▼
┌─────────────────────────────────────────┐
│            CLI (main.go)                │
│  1. Load config                         │
│  2. Fetch GTFS zip (HTTP/file)          │
│  3. Fetch GTFS-RT (HTTP)                │
│  4. Pass raw bytes to library           │
│  5. Format output                       │
│  6. Print to stdout                     │
└─────────────────────────────────────────┘
       │
       ▼ (raw bytes)
┌─────────────────────────────────────────┐
│         Core Library                    │
│  - Parse GTFS & GTFS-RT                 │
│  - Convert to SIRI                      │
│  - Return structured data               │
└─────────────────────────────────────────┘
```

## Output and Logs

The CLI outputs logs to **stderr** and SIRI data to **stdout**, allowing clean separation:

```bash
# Redirect output to file, logs to console
./gtfsrt-to-siri -call=vm > output.json

# Redirect both separately
./gtfsrt-to-siri -call=vm > output.json 2> logs.txt
```

## Error Handling

The CLI fails gracefully with clear error messages:

```bash
# Missing config
$ ./gtfsrt-to-siri
FATAL: Failed to load config: config.yml not found

# Invalid URL
$ ./gtfsrt-to-siri
FATAL: Failed to fetch GTFS: HTTP 404 Not Found

# Parse error
$ ./gtfsrt-to-siri
FATAL: Failed to create GTFS index: invalid zip format
```

## For Production Use

**Don't use the CLI for production systems.** Instead:

1. Integrate the library directly into your application
2. Implement your own data fetching (Kafka, MinIO, etc.)
3. Handle your own config management
4. Add proper error handling, retry logic, and monitoring
5. Cache the GTFS index in memory (parse once, reuse)

See main [README.md](../../README.md) for library integration examples.

## Performance

The CLI re-parses GTFS on every execution (oneshot), which is slow:
- GTFS parsing: 500ms-2s
- GTFS-RT parsing: 10-50ms
- Conversion: <1ms
- Formatting: 5-20ms

For high-throughput scenarios, use the library directly and cache the GTFS index.

