/*
Package gtfsrt provides data structures and utilities for working with GTFS-Realtime data.

This package is data-source agnostic - it accepts raw protobuf bytes and provides
convenient access methods. It does NOT handle HTTP fetching or file I/O.

# Basic Usage

Load from raw protobuf bytes:

	// Fetch protobuf bytes from your source (HTTP, Kafka, files, etc.)
	tuBytes := fetchTripUpdatesFromYourSource()
	vpBytes := fetchVehiclePositionsFromYourSource()
	saBytes := fetchServiceAlertsFromYourSource()

	// Create wrapper from raw bytes
	wrapper, err := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, saBytes)
	if err != nil {
	    log.Fatal(err)
	}

	// Access parsed data
	trips := wrapper.GetAllMonitoredTrips()
	for _, tripID := range trips {
	    routeID := wrapper.GetRouteIDForTrip(tripID)
	    vehicleRef := wrapper.GetVehicleRefForTrip(tripID)
	    fmt.Printf("Trip %s on route %s, vehicle %s\n", tripID, routeID, vehicleRef)
	}

# Protobuf Parsing

The wrapper automatically parses GTFS-RT protobuf messages and builds internal
indices for fast lookup. You can pass nil or empty byte slices for feeds you
don't have:

	// Only trip updates available
	wrapper, err := gtfsrt.NewGTFSRTWrapper(tuBytes, nil, nil)

	// Only vehicle positions
	wrapper, err := gtfsrt.NewGTFSRTWrapper(nil, vpBytes, nil)

# Data Access

Access methods provide convenient lookups without exposing protobuf internals:

	// Trip information
	routeID := wrapper.GetRouteIDForTrip(tripID)
	direction := wrapper.GetRouteDirectionForTrip(tripID)
	vehicleRef := wrapper.GetVehicleRefForTrip(tripID)

	// Stop time updates
	stops := wrapper.GetOnwardStopIDsForTrip(tripID)
	arrivalTime := wrapper.GetExpectedArrivalTimeAtStopForTrip(tripID, stopID)
	departureTime := wrapper.GetExpectedDepartureTimeAtStopForTrip(tripID, stopID)

	// Vehicle position
	lat, latOK := wrapper.GetVehicleLatForTrip(tripID)
	lon, lonOK := wrapper.GetVehicleLonForTrip(tripID)
	bearing, bearingOK := wrapper.GetVehicleBearingForTrip(tripID)

	// Service alerts
	alerts := wrapper.GetAlerts()

# Data Fetching

This package does NOT include HTTP fetching or file I/O. The library is purely for
parsing and accessing GTFS-RT data. Fetching logic belongs in your application layer.

For CLI usage, see cmd/gtfsrt-to-siri/ which includes a simple HTTP/file fetcher.
For server usage, implement your own fetching (Kafka, MinIO, etc.) and pass raw bytes
to NewGTFSRTWrapper.

# Supported Feed Types

- Trip Updates: Real-time arrival/departure predictions for scheduled trips
- Vehicle Positions: Current location, bearing, and status of vehicles
- Service Alerts: Disruptions, delays, and service changes
*/
package gtfsrt
