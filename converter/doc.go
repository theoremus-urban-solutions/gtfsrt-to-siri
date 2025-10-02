// Package converter is the main entry point for GTFS-Realtime to SIRI conversion.
//
// This package provides the core conversion logic that integrates GTFS static data,
// GTFS-Realtime feeds, and configuration to produce SIRI (Service Interface for
// Real-time Information) responses.
//
// # Overview
//
// The converter package coordinates three main components:
//   - GTFS static data (routes, stops, trips, shapes) via gtfs.GTFSIndex
//   - GTFS-Realtime feeds (trip updates, vehicle positions, alerts) via gtfsrt.GTFSRTWrapper
//   - Application configuration (field mutators, agency ID, etc.) via config.AppConfig
//
// # Usage
//
// Basic setup:
//
//	import (
//	    "mta/gtfsrt-to-siri/config"
//	    "mta/gtfsrt-to-siri/converter"
//	    "mta/gtfsrt-to-siri/formatter"
//	    "mta/gtfsrt-to-siri/gtfs"
//	    "mta/gtfsrt-to-siri/gtfsrt"
//	)
//
//	// Load configuration
//	cfg, _ := config.LoadAppConfig("config.yml")
//
//	// Initialize GTFS static data
//	gtfsIdx, _ := gtfs.NewGTFSIndexFromConfig(cfg.GTFS)
//
//	// Initialize GTFS-RT wrapper
//	rt := gtfsrt.NewGTFSRTWrapper(
//	    cfg.GTFSRT.TripUpdatesURL,
//	    cfg.GTFSRT.VehiclePositionsURL,
//	    cfg.GTFSRT.ServiceAlertsURL,
//	)
//
//	// Fetch latest real-time data
//	rt.Refresh()
//
//	// Create converter
//	conv := converter.NewConverter(gtfsIdx, rt, cfg)
//
// # Generating SIRI Responses
//
// The converter supports three SIRI modules:
//
// Vehicle Monitoring (VM) - Real-time vehicle positions and journey progress:
//
//	vmResp := conv.GetCompleteVehicleMonitoringResponse()
//	rb := formatter.NewResponseBuilder()
//	xmlBytes := rb.BuildXML(vmResp)
//
// Estimated Timetable (ET) - Predicted arrival/departure times:
//
//	et := conv.BuildEstimatedTimetable()
//	etResp := formatter.WrapEstimatedTimetableResponse(et, cfg.GTFS.AgencyID)
//	xmlBytes := rb.BuildXML(etResp)
//
//	// With filtering
//	filtered := formatter.FilterEstimatedTimetable(et, "STOP_123", "ROUTE_A", "0")
//
// Situation Exchange (SX) - Service alerts and disruptions:
//
//	sx := conv.BuildSituationExchange()
//	timestamp := rt.GetTimestampForFeedMessage()
//	sxResp := formatter.WrapSituationExchangeResponse(sx, timestamp, cfg.GTFS.AgencyID)
//	xmlBytes := rb.BuildXML(sxResp)
//
// # Architecture
//
// The converter package is organized into specialized files:
//   - converter.go: Main Converter struct and initialization
//   - vm.go: Vehicle Monitoring (SIRI-VM) conversion logic
//   - et.go: Estimated Timetable (SIRI-ET) conversion logic
//   - sx.go: Situation Exchange (SIRI-SX) conversion logic
//   - calls.go: Shared call/stop building utilities
//
// # Thread Safety
//
// Converter instances are NOT thread-safe. Each request should use its own converter
// instance or implement appropriate locking mechanisms. The underlying GTFS index and
// GTFS-RT wrapper can be safely shared across multiple converters.
//
// # Performance Considerations
//
// - The converter performs on-demand conversion with no internal caching
// - GTFS index is loaded once and kept in memory for fast lookups
// - GTFS-RT data should be refreshed periodically (e.g., every 30 seconds)
// - Vehicle position tracking uses snapshot-based estimation for missing data
// - Expected to handle 20MB+ XML outputs efficiently
//
// # Integration Example
//
// Typical server integration pattern:
//
//	// Initialization (once at startup)
//	gtfsIdx, _ := gtfs.NewGTFSIndexFromConfig(cfg.GTFS)
//	rt := gtfsrt.NewGTFSRTWrapper(...)
//
//	// Background refresh (every 30 seconds)
//	go func() {
//	    ticker := time.NewTicker(30 * time.Second)
//	    for range ticker.C {
//	        rt.Refresh()
//	    }
//	}()
//
//	// Per-request handling
//	http.HandleFunc("/siri/et", func(w http.ResponseWriter, r *http.Request) {
//	    conv := converter.NewConverter(gtfsIdx, rt, cfg)
//	    et := conv.BuildEstimatedTimetable()
//	    resp := formatter.WrapEstimatedTimetableResponse(et, cfg.GTFS.AgencyID)
//	    rb := formatter.NewResponseBuilder()
//	    w.Header().Set("Content-Type", "application/xml")
//	    w.Write(rb.BuildXML(resp))
//	})
package converter
