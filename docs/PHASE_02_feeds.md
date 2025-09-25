# Phase 02 — Feeds: GTFS Static Indices and GTFS-Realtime Reader

Goal: load static GTFS indices and implement GTFS-RT protobuf reader + wrapper API compatible with Node usage.

## GTFS Static (feed_gtfs.go)
- [ ] Load indexedScheduleData.json & indexedSpatialData.json
- [ ] Implement accessors used by Node code:
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

## Health & Observability
- [ ] Expose last successful GTFS-RT timestamp in health

## References
- Node: lib/ConverterStream.js, lib/converter/*.js, lib/trainTracking/*.js
- Server repo: proto_files/, GTFSRealtime_FeedReaderService.js
