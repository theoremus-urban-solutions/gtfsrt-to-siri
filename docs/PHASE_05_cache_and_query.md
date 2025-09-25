# Phase 5 — Cache, Buffering, Query Selection

Goal: pre-buffer journeys and calls for fast assembly; implement query processing and error responses.

## cache.go
- [ ] Build existence maps for stops/routes (lowercased keys; apply mutators when needed)
- [ ] Maintain filters from GTFS-RT wrapper (trains/routes/stops with alerts)
- [ ] Store converter, GTFS, GTFSrt references; unscheduledTripIndicator
- [ ] Buffer building order: Calls → MVJs → SituationExchange (calls mutates objects in Node; mirror effects safely)
- [ ] Response memoization map with key `[tripKeys|deliveryType|detail|maxCalls|stop|includeSX|format]`
- [ ] GetState(): return serialized converter state

## Bufferers
- [ ] bufferMonitoredVehicleJourneys: pack journeys into JSON/XML byte blocks; create byTripKeysIndex offsets; partition by route/direction; map vehicleRef→gtfsTripKey
- [ ] bufferCalls: pack calls into JSON/XML byte blocks; indices for stopID→callNumber, ETA sorting per stop; per-route and per-direction partitions
- [ ] bufferSituationExchange: stringified JSON and XML bytes

## query.go
- [ ] Parse case-insensitive params and keep originals under _param keys
- [ ] Validate:
  - [ ] monitoringref required (SM)
  - [ ] directionref in {0,1}
  - [ ] operatorref matches agency IDs
  - [ ] stop/route existence
- [ ] Selection logic:
  - [ ] VehicleMonitoring: vehicleref→tripKey or by route/direction; apply MaximumStopVisits and MinimumStopVisitsPerLine rules (ETA ordered fill)
  - [ ] StopMonitoring: by stop and route/direction using indices; same limits behavior
- [ ] BuildErrorResponse: construct SIRI ErrorCondition JSON/XML

## References
- Node: lib/caching/ConverterCache.js, CachedMessageBufferers.js, QueryProcessor.js
