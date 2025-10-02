// Package tracking provides real-time vehicle position tracking and estimation.
//
// This package handles:
// - Capturing vehicle positions from GTFS-RT at specific timestamps
// - Estimating vehicle locations along routes using GTFS shapes
// - Projecting vehicles onto their route geometry
// - Tracking bearing/heading and distance traveled
//
// The Snapshot type represents a point-in-time capture of all vehicle positions,
// which can be used for historical position estimation when real-time data is missing.
package tracking
