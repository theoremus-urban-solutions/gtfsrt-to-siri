## Go Requirements for GTFS-Realtime → SIRI Converter Service

This document defines what must be implemented to provide a production-grade Go service that ingests MTA Subway GTFS-Realtime, converts to SIRI (VehicleMonitoring and StopMonitoring) and serves JSON/XML responses. It follows our flat Go service guidelines (see ../guidelines.md). Project root for Go code is ./golang.

### 1) Service Category and Scope
- Type: Data Ingest + API
  - Pull GTFS static (indexed JSON) and GTFS-RT protobuf; transform to SIRI; expose HTTP endpoints.
- Non-goals: Deep SIRI alert semantics; hosted admin console UI. Aim for feature parity with Node converter responses.

### 2) Project Layout (flat, under package main)
- ./golang/
  - main.go: boot, shutdown, http server
  - config.go: YAML config structs + validation
  - feed_gtfs.go: load static GTFS indices (schedule/spatial)
  - feed_gtfsrt.go: GTFS-RT protobuf reader + wrapper
  - converter.go: high-level conversion orchestrator
  - vm_builder.go: MonitoredVehicleJourney builder
  - call_builder.go: MonitoredCall / OnwardCalls builder
  - sx_builder.go: SituationExchange builder (basic)
  - cache.go: response buffer cache + memoization
  - response.go: JSON/XML assembly (buffers), timestamps, ValidUntil
  - query.go: params parsing/validation, trip key selection
  - tracker.go: train tracker snapshot, inference
  - tracker_utils.go: geometry ops, bearings
  - presentable_distance.go: rules implementation
  - handler_monitoring.go: HTTP handlers for SIRI endpoints
  - health.go: health and metrics
  - auth.go: API key authorization (+ open access mode)
  - logging.go: logging init
  - vendor libs: via go.mod
- CLI entrypoint (standard Go practice) while keeping a flat library layout:
  - ./golang/cmd/gtfsrt-to-siri/main.go: thin CLI that wires config + feeds and prints a one-shot SIRI response or runs the server mode (flags). The rest of the code remains flat under package main.

### 3) Config (config.yml + struct tags)
- Sections:
  - server: port, timeouts, auth mode (openAccess|keyed), rate limits
  - gtfs: paths to indexedScheduleData.json, indexedSpatialData.json, agency_id
  - gtfsrt: feedURL, readIntervalMS, request timeouts
  - converter:
    - fieldMutators: OriginRef/StopPointRef/DestinationRef `[from, to]`
    - unscheduledTripIndicator (string)
    - callDistanceAlongRouteNumOfDigits (int)
- Validation: use github.com/go-playground/validator per guidelines.

### 4) Inputs and Tooling
- GTFS static: same indexed JSON shape used in Node (must replicate accessors used by Node wrappers):
  - Route/Trip lookup; origin/destination stop; shapes; stop distances; previous/next stop; agency timezone.
- GTFS-RT: parse protobuf (`gtfs-realtime.proto`, `nyct-subway.proto`), provide wrapper methods used by Node code:
  - Examples to support: `GetAllMonitoredTrips()`, `GetGTFSTripKeyForRealtimeTripKey()`, `GetRouteIDForTrip()`, `GetRouteDirectionForTrip()`, `GetOnwardStopIDsForTrip()`, `GetExpectedArrivalTimeAtStopForTrip()`, `GetVehiclePositionTimestamp()`, `GetTimestampForFeedMessage()`, `GetStartDateForTrip()`.

### 5) Conversion Semantics (feature parity targets)
- VehicleMonitoringDelivery:
  - For each realtime trip_id, build MVJ with fields as in Node `MonitoredVehicleJourneyBuilder.js` (LineRef, DirectionRef, FramedVehicleJourneyRef, JourneyPatternRef, PublishedLineName, OperatorRef, OriginRef, DestinationRef/Name, OriginAimedDepartureTime, SituationRef=null, Monitored=true, VehicleLocation, Bearing, ProgressRate=null, ProgressStatus=null, VehicleRef, OnwardCalls).
- StopMonitoringDelivery:
  - Build MonitoredStopVisit entries with the journey plus MonitoredCall and OnwardCalls semantics below.
- Calls:
  - StopMonitoring: always include MonitoredCall; include OnwardCalls up to MaximumNumberOfCallsOnwards; must ensure inclusion up to and including selected stop if limit is lower than stop index.
  - VehicleMonitoring: include calls only if detailLevel=="calls" and max>0.
- SituationExchangeDelivery: include only when affected trains/routes/stops exist per alert filter objects; content may be simplified to generic text.
- Timestamps/ValidUntil:
  - Two ResponseTimestamp insertions; ValidUntil = feedTimestamp + readIntervalMS (or null).

### 6) Query Parameters and Validation
- StopMonitoring: monitoringref (required), lineref, directionref in {0,1}, stopmonitoringdetaillevel ("calls"|"normal"), maximumnumberofcallsonwards, maximumstopvisits, minimumstopvisitsperline, operatorref.
- VehicleMonitoring: vehicleref, lineref, directionref, vehiclemonitoringdetaillevel, maximumnumberofcallsonwards, maximumstopvisits, minimumstopvisitsperline, operatorref.
- Lowercase parsing plus raw value mirrors (e.g. keep _param for original case) to preserve current behavior.
- Existence checks:
  - operatorref is a valid GTFS agency_id; route and stop exist in index.
- Error responses:
  - Build SIRI ErrorCondition JSON/XML compatible with current Node `QueryError` structure.

### 7) Response Assembly Strategy (Performance)
- Assemble JSON/XML using pre-sized `[]byte` buffers and `copy()` operations to avoid allocations for each element; compute lengths first, allocate once, then fill slices sequentially (parity with Node buffers).
- Memoize responses keyed by: `tripKeySet|deliveryType|detail|maxCalls|stop|includeSX|format`.
- Timestamps are copied into predetermined offsets after initial buffer creation.

### 8) Train Tracking (Snapshot + Inference)
- Snapshot per feed timestamp; reject older GTFS-RT than last snapshot.
- Invariants:
  - LineString always spans current inferred position to the immediate next stop.
  - Single previous snapshot retained.
  - Stable GTFS dataset tied to each tracked trip.
- States:
  - FIRST_OCCURRENCE, KNEW_LOCATION, AT_STOP, AT_ORIGIN, AT_DESTINATION, NO_ETA, BAD_PREVIOUS_ETA, OUT_OF_SEQUENCE_STOPS, AT_INTERMEDIATE_STOP, HAS_MOVED, SAME_IMMEDIATE_NEXT_STOP, NO_STOP_TIME_UPDATE.
- Movement:
  - At stop: Point or LineString to subsequent stop with atStop=true.
  - Between stops: advance along LineString by ratio of elapsed/eta delta; extend path if endpoint advances; maintain bearing.
- PresentableDistance: implement D=0.5mi, N=3, E=0.5mi, P=500ft, T=100ft.

### 9) Validators (new, Go-side)
- GTFS-RT validation options:
  - CI/offline: integrate MobilityData gtfs-realtime-validator Java tool in pipeline.
  - Runtime (optional, dev flag): basic sanity checks (monotonic timestamps, required message fields, allowed ranges).
- SIRI validation options:
  - XML: optional XSD validation (xmllint or Go XML schema lib) in dev/CI; turn off in prod.
  - JSON: define JSON Schemas mirroring our SIRI outputs; validate with `github.com/xeipuuv/gojsonschema` (dev only).

### 10) HTTP API
- Routes:
  - GET /api/siri/vehicle-monitoring.(json|xml)
  - GET /api/siri/stop-monitoring.(json|xml)
  - GET /api/health
  - GET /metrics (optional)
- Auth:
  - OpenAccess or API key check; 429 via toobusy-like rate limit.
- Responses: stream `[]byte` payloads; set correct Content-Type, Content-Length.

### 11) Observability
- Logging: structured logs, request ID.
- Health: check GTFS reader state, last feed timestamp recency, GTFS index loaded.
- Metrics: counters for requests by type/format, response sizes, conversion latency; gauge for latest GTFS-RT timestamp.
- Tracing: optional via shared lib.

### 12) Operational Concerns
- Hot reload: support config reload without restart for feed URLs and converter config.
- Memory: cap memoization size or use LRU; monitor allocations due to buffer assembly.
- Error handling: never panic on single request; emit SIRI error payloads.

### 13) Testing
- Unit: builders, query selection, response assembly offsets/lengths, distance rules.
- Integration: replay sample GTFS and GTFS-RT messages and diff against Node outputs.
- Load: basic concurrency tests; ensure no data races.

### 14) Migration / Parity Checklist
- Achieve field-by-field parity against Node for representative trips and stops, both JSON and XML.
- Validate parameter handling and error messages.
- Confirm memoization and timestamps/ValidUntil offsets.

