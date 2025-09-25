# Phase 4 — SIRI Builders

Goal: implement MonitoredVehicleJourney, Calls, and SituationExchange builders matching Node semantics. (Library scope.)

## vm_builder.go (MonitoredVehicleJourney)
- [ ] Build MVJ for each realtime trip
  - [ ] LineRef(agency_id + '_' + route_id when agency present)
  - [ ] DirectionRef (GTFS direction_id)
  - [ ] FramedVehicleJourneyRef { DataFrameRef(YYYY-MM-DD), DatedVehicleJourneyRef }
  - [ ] JourneyPatternRef (agency_id + '_' + shape_id)
  - [ ] PublishedLineName (route_short_name)
  - [ ] OperatorRef (agency_id)
  - [ ] OriginRef/DestinationRef (mutators applied; GTFS-RT last stoptime fallback)
  - [ ] DestinationName (trip_headsign)
  - [ ] OriginAimedDepartureTime (startDate + scheduled or origin time if provided)
  - [ ] SituationRef (null)
  - [ ] Monitored (true)
  - [ ] VehicleLocation (from tracker)
  - [ ] Bearing (from tracker)
  - [ ] ProgressRate/ProgressStatus (null)
  - [ ] VehicleRef (agency_id + '_' + train_id)
  - [ ] OnwardCalls (from CallBuilder)

## call_builder.go (Calls)
- [ ] Build MonitoredCall + OnwardCalls
  - [ ] Distances: PresentableDistance, DistanceFromCall, StopsFromCall, CallDistanceAlongRoute
  - [ ] ExpectedArrivalTime/ExpectedDepartureTime (ISO8601)
  - [ ] StopPointRef (mutator), StopPointName
  - [ ] VisitNumber (1)
  - [ ] Respect StopMonitoring semantics: always include MonitoredCall; ensure inclusion up to selected stop when limiting calls
  - [ ] VM semantics: include calls only when detail=="calls" and max>0

## sx_builder.go (SituationExchange)
- [ ] Include only when filters indicate alerts (affected trains/routes/stops)
- [ ] Minimal PtSituationElement (Severity=undefined, Summary/Description generic)
- [ ] Affects → VehicleJourneys by route/direction

## References
- Node: lib/converter/MonitoredVehicleJourneyBuilder.js, CallBuilder.js, SituationExchangeDeliveryBuilder.js
