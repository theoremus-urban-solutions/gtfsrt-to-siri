# Phase 02 — Feeds: GTFS Static (ZIP) and GTFS-Realtime Reader

Goal: ingest GTFS static from a ZIP and implement GTFS-RT protobuf reader + wrapper API compatible with Node usage.

## GTFS Static (feed_gtfs.go)
- [ ] Ingest GTFS ZIP (routes.txt, trips.txt, stops.txt, stop_times.txt, agency.txt, shapes.txt)
- [ ] Build in-memory indices used by Node code:
  - [ ] GetAgencyTimezone(agencyID)
  - [ ] getOriginStopIdForTrip(gtfsTripKey)
  - [ ] getDestinationStopIdForTrip(gtfsTripKey)
  - [ ] getTripHeadsign(gtfsTripKey)
  - [ ] getShapeIDForTrip(gtfsTripKey)
  - [ ] getFullTripIDForTrip(gtfsTripKey)
  - [ ] getBlockIDForTrip(gtfsTripKey)
  - [ ] getRouteShortName(routeID)
  - [ ] getStopName(stopID)
  - [ ] Distances/geometry: getStopDistanceAlongRouteForTripInMeters/Kilometers, getPreviousStopIDOfStopForTrip, getSliceShapeForTrip, getSnappedCoordinatesOfStopForTrip, getShapeSegmentNumberOfStopForTrip
  - [ ] Flags: tripIsAScheduledTrip, tripsHasSpatialData
  - [ ] Enumerations: getAllStops(), getAllRoutes(), getAllAgencyIDs()

## GTFS-Realtime (feed_gtfsrt.go)
- [ ] Integrate protobufs (gtfs-realtime.proto, nyct-subway.proto)
- [ ] Implement reader with configurable feedURL and timeouts
- [ ] Wrapper methods:
  - [ ] getAllMonitoredTrips()
  - [ ] getGTFSTripKeyForRealtimeTripKey(tripID)
  - [ ] getRouteIDForTrip(tripID), getRouteDirectionForTrip(tripID)
  - [ ] getOnwardStopIDsForTrip(tripID)
  - [ ] getExpectedArrivalTimeAtStopForTrip(tripID, stopID), getExpectedDepartureTimeAtStopForTrip(...)
  - [ ] getIndexOfStopInStopTimeUpdatesForTrip(tripID, stopID)
  - [ ] getStartDateForTrip(tripID), getOriginTimeForTrip(tripID)
  - [ ] getVehiclePositionTimestamp(tripID), getTimestampForTrip(tripID), getTimestampForFeedMessage()
  - [ ] Alerts: getAllTripsWithAlert(), getTrainsWithAlertFilterObject(), getStopsWithAlertFilterObject(), getRoutesWithAlertFilterObject()
- [ ] Timezone: set agency timezone for time formatting

## References
- Node: lib/ConverterStream.js, lib/converter/*.js, lib/trainTracking/*.js
