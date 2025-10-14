// Package siri defines SIRI (Service Interface for Real-time Information) data types.
//
// SIRI is a European standard (CEN/TS 15531) for real-time public transport information.
// This package contains Go structs for three SIRI modules:
//
//   - VehicleMonitoringDelivery (VM): Real-time vehicle locations and status
//   - EstimatedTimetable (ET): Complete journey timetables with predictions
//   - SituationExchangeDelivery (SX): Service alerts and disruptions
//
// All types include JSON and XML struct tags for serialization.
package siri
