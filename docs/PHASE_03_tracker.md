# Phase 3 — Train Tracker Snapshot & Inference

Goal: implement snapshot lifecycle, state machine, location inference, and presentable distance. (Library scope only; no HTTP dependencies.)

## Snapshot (tracker.go)
- [ ] newSnapshot(GTFS, GTFSrt, config, initialState)
  - [ ] Reject older GTFS-RT vs previous snapshot
  - [ ] Return same snapshot if timestamps equal
  - [ ] Create Snapshot with gtfsrtTimestamp, previousSnapshot=null link, tripKeyToGTFSVersion map
- [ ] Getters: GetLatitude, GetLongitude, GetBearing, GetVehicleDistanceAlongRouteInKilometers, GetDistanceFromCall

## Inference (tracker_utils.go)
- [ ] getImmediateStopInfo: stopId, timestamp, eta, atStop, sequenceNumber, distance_along_route_in_km
- [ ] getLineStringBetweenStopsForTrip: build LineString with properties (gtfsTripKey, end_stop_id, bearing, start_dist_along_route_in_km, line_distance_km)
- [ ] getGeoJSONPointForStopForTrip: Point with properties
- [ ] advancePositionAlongLineString: verify params; compute kilometersCovered; along; slice; update bearing; start_dist; line_distance
- [ ] extendLinestringToFurtherStopForTrip: verify; append waypoints; recompute line_distance; penultimate_stop_id
- [ ] getMTABearing: convert turf bearing to MTA convention (0 east, CCW)

## State Machine (tracker.go)
- [ ] getStateOfTrain: compute flags FIRST_OCCURRENCE, KNEW_LOCATION, AT_STOP, AT_ORIGIN, AT_DESTINATION, NO_ETA, BAD_PREVIOUS_ETA, OUT_OF_SEQUENCE_STOPS, AT_INTERMEDIATE_STOP, HAS_MOVED, SAME_IMMEDIATE_NEXT_STOP, NO_STOP_TIME_UPDATE
- [ ] inferTrainLocation: implement control flow mirroring Node (reuse/recompute geometry; handle errors; emit diagnostics)

## Presentable Distance (presentable_distance.go)
- [ ] Implement D=0.5mi, N=3, E=0.5mi, P=500ft, T=100ft rules

## Diagnostics
- [ ] Expose tracker state via converter GetState()
- [ ] Hook logging for anomalies and errors (no panics)

## References
- Node: lib/trainTracking/TrainTracker.js, InferenceEngine.js, Utils.js, PresentableDistanceCalculator.js, Constants.js
