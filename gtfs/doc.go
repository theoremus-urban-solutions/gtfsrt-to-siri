/*
Package gtfs provides GTFS static data loading and indexing.

This package is data-source agnostic - it accepts raw zip bytes or io.Reader
and builds an in-memory index. It does NOT handle HTTP downloads or file paths.

# Basic Usage

Load from raw bytes:

	// Fetch GTFS zip from your source (HTTP, MinIO, files, etc.)
	gtfsZipBytes := fetchGTFSFromYourSource()

	// Create index from bytes
	index, err := gtfs.NewGTFSIndexFromBytes(gtfsZipBytes, "AGENCY_ID")
	if err != nil {
	    log.Fatal(err)
	}

	// Access indexed data
	routeID := index.GetRouteIDForTrip("trip_123")
	stopName := index.GetStopName("stop_456")

Load from io.Reader:

	// Open GTFS zip from any source
	file, _ := os.Open("gtfs.zip")
	defer file.Close()
	stat, _ := file.Stat()

	// Create index from reader
	index, err := gtfs.NewGTFSIndexFromReader(file, stat.Size(), "AGENCY_ID")

# Performance: Cache the Index

⚠️ IMPORTANT: Parse GTFS once at startup and keep the index in memory.
GTFS is static data - parsing the zip on every request is wasteful (500ms-2s vs <1ms).

	// Server initialization - parse once
	gtfsBytes, _ := fetchFromMinIO("gtfs.zip")
	cachedIndex, _ := gtfs.NewGTFSIndexFromBytes(gtfsBytes, "AGENCY")

	// Per-request - reuse cached index (fast!)
	conv := converter.NewConverter(cachedIndex, rt, opts)

# Data Structure

The index provides fast lookups for:

- Routes (route_id → route_short_name, route_type)
- Stops (stop_id → stop_name, lat/lon)
- Trips (trip_id → route_id, headsign, direction)
- Stop sequences (trip_id → ordered list of stop_ids)
- Stop times (trip_id + stop_id → arrival/departure time)
- Shapes (shape_id → ordered list of lat/lon points)

# Agency ID

Agency ID is required for proper SIRI reference formatting:

- LineRef: {agency_id}:Line:{route_id}
- StopPointRef: {agency_id}:Quay:{stop_id}
- VehicleRef: {agency_id}:VehicleRef:{vehicle_id}

# Downloading GTFS

This package includes a FetchGTFSData() helper for CLI usage.
For server usage, implement your own fetching logic and pass bytes to NewGTFSIndexFromBytes.

# Memory Footprint

Typical GTFS index: ~100-200MB for 10-20K stops and 1K routes.
*/
package gtfs
