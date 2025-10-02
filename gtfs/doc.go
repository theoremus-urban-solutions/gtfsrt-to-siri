// Package gtfs handles loading, parsing, and indexing GTFS static feeds.
//
// It fetches GTFS ZIP files from URLs, parses CSV files (stops, routes, trips,
// shapes), and builds an in-memory index for fast lookups.
//
// The main type is GTFSIndex which provides accessor methods for GTFS data.
package gtfs
