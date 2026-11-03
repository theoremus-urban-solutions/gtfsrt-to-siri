# Changelog

Every release of gtfsrt-to-siri, newest first. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow [Semantic Versioning](https://semver.org/).

## [1.1.0] - 2026-11-03

This release is about memory. The realtime-transit service kept running out of it on staging, and a large share went through this library. After moving to 1.1.0, the service's peak memory on staging fell from 677 MB to 285 MB.

### Changed

- **The XML escaper is built once.** It used to be rebuilt for every string it escaped, which accounted for about three quarters of all memory a conversion allocated. A Sofia Vehicle Monitoring conversion now allocates 6.4 MB instead of 61.8 MB and takes 12 ms instead of 44 ms. A BDZ Estimated Timetable conversion allocates 4.5 MB instead of 48.0 MB.
- **GTFS files are read row by row.** The loader no longer reads a whole CSV file into memory before indexing it.
- **Repeated values are stored once.** Stop IDs, trip IDs and times that appear many times in the GTFS share a single copy instead of each holding on to the line it was read from.
- **One stop-time list per trip.** The index keeps a single ordered list of stop times per trip instead of several maps. Together with the two changes above, Sofia's timetable keeps 48 MB in memory instead of 155 MB, allocates 218 MB while loading instead of 548 MB, and loads in 0.93 s instead of 1.61 s.
- SIRI output is otherwise unchanged. It was compared element by element on eight live feeds from five datasets before and after.

### Fixed

- **`ValidUntilTime` is written again** on every `VehicleActivity` in Vehicle Monitoring. SIRI 2.1 requires it, and 1.0.7 had stopped writing it. It is the vehicle's timestamp plus `ConverterOptions.ReadIntervalMS`, which is back as an option. When the option is zero, `DefaultReadIntervalMS` (30 seconds) is used.

### Upgrading

- `GTFSIndex.TripStopIdx` and `GTFSIndex.StopTimes` are removed. Stop times now live in `GTFSIndex.TripStopTimes`, in the same order as `TripStopSeq`. The lookup methods `GetArrivalTime`, `GetDepartureTime`, `GetPickupType`, `GetDropOffType` and `GetPreviousStopIDOfStopForTrip` return the same results as before, including for trips that visit a stop twice.
- `DeserializeIndex`, `DeserializeIndexFromReader` and `DeserializeIndexFromFile` return an error for an index cached by an older version, instead of loading it without stop times. Rebuild the index from the GTFS zip and cache it again.
- Coming from 1.0.6: `FieldMutators` was removed in 1.0.7 and stays removed.

## [1.0.7] - 2025-11-06

### Changed

- **Simpler command line.** The CLI now takes feed URLs directly with `-tripUpdates`, `-vehiclePositions` and `-alerts`, or reads them from `config.yml` with `-read-from-config`. It produces every SIRI module the given feeds allow, and `-format` defaults to `xml`.
- `config.yml` keeps only the server port, the GTFS static URL, the agency ID and the GTFS-RT feed URLs.
- The README was rewritten, and a `CLAUDE.md` guide for coding agents was added.

### Removed

- CLI flags `-mode`, `-call`, `-feed`, `-serviceAlerts` (now `-alerts`), `-monitoringRef`, `-lineRef`, `-directionRef` and `-modules`.
- `config.yml` settings `readIntervalMS`, `timeoutMS` and the whole `converter` section (`fieldMutators`, `unscheduledTripIndicator`, `callDistanceAlongRouteNumOfDigits`, `tripKeyStrategy`).
- `ConverterOptions.FieldMutators` and the `FieldMutators` type, which rewrote stop, origin and destination references.
- `ConverterOptions.ReadIntervalMS`, which also removed `ValidUntilTime` from Vehicle Monitoring. Both came back in 1.1.0.

## [1.0.6] - 2025-10-28

### Added

- `NewConverterWithCachedGTFS`, which reuses an already loaded `GTFSIndex` instead of loading the GTFS for every conversion.
- Saving and loading a built index: `SerializeIndex`, `DeserializeIndex`, and their file and stream variants, so a service can cache the index on disk.

### Changed

- SIRI types moved to the shared [transit-types](https://github.com/theoremus-urban-solutions/transit-types) library, so other services use the same structures. The local `siri` type definitions were removed.
- `GTFSIndex` fields are exported so the index can be saved, and each trip has a stop position lookup that speeds up conversions.

### Fixed

- `VehicleMode` now shows trams correctly.

## [1.0.5] - 2025-10-27

### Changed

- **Clearer warnings.** Data quality problems are collected during a conversion and logged once per kind, with a count, up to three example IDs, what is missing and what the converter did instead.

## [1.0.4] - 2025-10-23

### Added

- Estimated Timetable builds calls from the GTFS-RT data alone when a trip is missing from the static GTFS, and uses realtime times where static times are missing.
- Situation Exchange falls back to the static GTFS when the realtime feed has no route ID. Vehicle Monitoring uses `UNKNOWN` when neither has one.
- Warnings for alerts without a summary or description, unknown stops, and vehicles whose origin stop has no name.

## [1.0.3] - 2025-10-21

### Changed

- **Lower memory use.** `shapes.txt` is no longer loaded, index structures are smaller, and XML output is no longer pretty-printed.
- Test feeds use the `.pbf` extension.

## [1.0.2] - 2025-10-12

### Added

- Situation Exchange carries summaries and descriptions in every language the alert provides, plus info links.

### Changed

- Situation Exchange `<Summary>` text is shorter and plainer.

### Fixed

- Trips that only appear in trip updates no longer show up in Vehicle Monitoring. It now lists only vehicles with a position.

## [1.0.1] - 2025-10-07

### Added

- The CLI accepts local `.pbf` files as well as URLs.

### Removed

- The HTTP client in the `gtfsrt` package. The library only parses bytes, and fetching belongs to the application.

## [1.0.0] - 2025-10-06

### Added

- First release. Converts raw GTFS static zips and GTFS-RT protobuf bytes into SIRI Vehicle Monitoring, Estimated Timetable and Situation Exchange, in XML or JSON, following the Nordic SIRI profile.
- Packages for GTFS loading, GTFS-RT parsing, conversion, formatting and vehicle tracking, a command-line tool, and a test suite on Sofia data.

[1.1.0]: https://github.com/theoremus-urban-solutions/gtfsrt-to-siri/compare/v1.0.7...v1.1.0
[1.0.7]: https://github.com/theoremus-urban-solutions/gtfsrt-to-siri/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/theoremus-urban-solutions/gtfsrt-to-siri/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/theoremus-urban-solutions/gtfsrt-to-siri/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/theoremus-urban-solutions/gtfsrt-to-siri/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/theoremus-urban-solutions/gtfsrt-to-siri/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/theoremus-urban-solutions/gtfsrt-to-siri/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/theoremus-urban-solutions/gtfsrt-to-siri/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/theoremus-urban-solutions/gtfsrt-to-siri/releases/tag/v1.0.0
