// Package gtfsrt handles fetching and indexing GTFS-Realtime protobuf feeds.
//
// It supports three feed types:
//   - Trip Updates: real-time arrival/departure predictions
//   - Vehicle Positions: current vehicle locations
//   - Service Alerts: disruptions and service changes
//
// The main type is GTFSRTWrapper which fetches feeds and provides indexed access.
package gtfsrt
