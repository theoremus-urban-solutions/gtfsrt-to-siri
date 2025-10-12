package gtfsrt

import (
	"fmt"
	"time"

	gtfsrtpb "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

// GTFSRTWrapper stores GTFS-Realtime data in memory for fast lookups.
// This wrapper is data-source agnostic - it accepts raw protobuf bytes
// and does NOT handle HTTP fetching or file I/O.
type GTFSRTWrapper struct {
	trips           map[string]struct{} // All trips (from both TripUpdates and VehiclePositions)
	tripsFromTU     map[string]struct{} // Trips from TripUpdates only (for ET)
	tripsFromVP     map[string]struct{} // Trips from VehiclePositions only (for VM)
	vehicleTS       map[string]int64
	headerTimestamp int64

	tripRoute      map[string]string           // trip_id -> route_id
	tripDir        map[string]string           // trip_id -> direction (string)
	tripDate       map[string]string           // trip_id -> start_date (YYYYMMDD)
	onwardStops    map[string][]string         // trip_id -> ordered stop_ids
	etaByStop      map[string]map[string]int64 // trip_id -> stop_id -> arrival epoch
	etdByStop      map[string]map[string]int64 // trip_id -> stop_id -> departure epoch
	schedRelByStop map[string]map[string]int32 // trip_id -> stop_id -> schedule_relationship (0=SCHEDULED, 1=SKIPPED, etc.)

	tripVehicleRef map[string]string  // trip_id -> vehicle id
	tripLat        map[string]float64 // trip_id -> lat
	tripLon        map[string]float64 // trip_id -> lon
	tripBearing    map[string]float64 // trip_id -> bearing
	tripSpeed      map[string]float64 // trip_id -> speed (m/s)

	// Occupancy and congestion data
	tripOccupancy  map[string]int32 // trip_id -> occupancy_status (from TripUpdate)
	tripCongestion map[string]int32 // trip_id -> congestion_level (from VehiclePosition)

	// Alerts data (parsed from GTFS-RT Alerts)
	alerts        []RTAlert
	alertsByRoute map[string][]int // route_id -> indices in alerts slice
	alertsByStop  map[string][]int // stop_id -> indices
	alertsByTrip  map[string][]int // trip_id -> indices
}

// NewGTFSRTWrapper creates a new wrapper from raw GTFS-RT protobuf bytes.
// Pass nil or empty byte slices for feeds you don't have.
//
// Example:
//
//	tuBytes := fetchTripUpdates() // From HTTP, Kafka, files, etc.
//	vpBytes := fetchVehiclePositions()
//	saBytes := fetchServiceAlerts()
//	wrapper, err := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, saBytes)
func NewGTFSRTWrapper(tripUpdatesData, vehiclePositionsData, serviceAlertsData []byte) (*GTFSRTWrapper, error) {
	wrapper := &GTFSRTWrapper{
		trips:          map[string]struct{}{},
		tripsFromTU:    map[string]struct{}{},
		tripsFromVP:    map[string]struct{}{},
		vehicleTS:      map[string]int64{},
		schedRelByStop: map[string]map[string]int32{},
		tripRoute:      map[string]string{},
		tripDir:        map[string]string{},
		tripDate:       map[string]string{},
		onwardStops:    map[string][]string{},
		etaByStop:      map[string]map[string]int64{},
		etdByStop:      map[string]map[string]int64{},
		tripVehicleRef: map[string]string{},
		tripLat:        map[string]float64{},
		tripLon:        map[string]float64{},
		tripBearing:    map[string]float64{},
		tripSpeed:      map[string]float64{},
		tripOccupancy:  map[string]int32{},
		tripCongestion: map[string]int32{},
		alerts:         []RTAlert{},
		alertsByRoute:  map[string][]int{},
		alertsByStop:   map[string][]int{},
		alertsByTrip:   map[string][]int{},
	}

	// Parse trip updates
	if len(tripUpdatesData) > 0 {
		var tuFeed gtfsrtpb.FeedMessage
		if err := proto.Unmarshal(tripUpdatesData, &tuFeed); err != nil {
			return nil, fmt.Errorf("failed to parse trip updates: %w", err)
		}
		wrapper.parseTripUpdatesFeed(&tuFeed)
	}

	// Parse vehicle positions
	if len(vehiclePositionsData) > 0 {
		var vpFeed gtfsrtpb.FeedMessage
		if err := proto.Unmarshal(vehiclePositionsData, &vpFeed); err != nil {
			return nil, fmt.Errorf("failed to parse vehicle positions: %w", err)
		}
		wrapper.parseVehiclePositionsFeed(&vpFeed)
	}

	// Parse service alerts
	if len(serviceAlertsData) > 0 {
		var saFeed gtfsrtpb.FeedMessage
		if err := proto.Unmarshal(serviceAlertsData, &saFeed); err != nil {
			return nil, fmt.Errorf("failed to parse service alerts: %w", err)
		}
		wrapper.parseServiceAlertsFeed(&saFeed)
	}

	// Set timestamp from first available feed
	if wrapper.headerTimestamp == 0 {
		wrapper.headerTimestamp = time.Now().Unix()
	}

	return wrapper, nil
}

// Accessor methods

func (w *GTFSRTWrapper) GetAllMonitoredTrips() []string {
	ids := make([]string, 0, len(w.trips))
	for id := range w.trips {
		ids = append(ids, id)
	}
	return ids
}

// GetTripsFromTripUpdates returns trips that have TripUpdate data (for Estimated Timetable)
func (w *GTFSRTWrapper) GetTripsFromTripUpdates() []string {
	ids := make([]string, 0, len(w.tripsFromTU))
	for id := range w.tripsFromTU {
		ids = append(ids, id)
	}
	return ids
}

// GetTripsFromVehiclePositions returns trips that have VehiclePosition data (for Vehicle Monitoring)
func (w *GTFSRTWrapper) GetTripsFromVehiclePositions() []string {
	ids := make([]string, 0, len(w.tripsFromVP))
	for id := range w.tripsFromVP {
		ids = append(ids, id)
	}
	return ids
}

func (w *GTFSRTWrapper) GetGTFSTripKeyForRealtimeTripKey(tripID string) string { return tripID }

// TripKeyForConverter returns a composite trip key for GTFS static lookups.
// Format: {agency}_{startDate}_{tripID} (or just tripID if agency/startDate empty)
func TripKeyForConverter(tripID, agency, startDate string) string {
	// Default strategy: combine all available parts
	key := tripID
	if startDate != "" {
		key = startDate + "_" + key
	}
	if agency != "" {
		key = agency + "_" + key
	}
	return key
}

func (w *GTFSRTWrapper) GetRouteIDForTrip(tripID string) string         { return w.tripRoute[tripID] }
func (w *GTFSRTWrapper) GetRouteDirectionForTrip(tripID string) string  { return w.tripDir[tripID] }
func (w *GTFSRTWrapper) GetOnwardStopIDsForTrip(tripID string) []string { return w.onwardStops[tripID] }

func (w *GTFSRTWrapper) GetExpectedArrivalTimeAtStopForTrip(tripID, stopID string) int64 {
	if m := w.etaByStop[tripID]; m != nil {
		return m[stopID]
	}
	return 0
}

func (w *GTFSRTWrapper) GetExpectedDepartureTimeAtStopForTrip(tripID, stopID string) int64 {
	if m := w.etdByStop[tripID]; m != nil {
		return m[stopID]
	}
	return 0
}

func (w *GTFSRTWrapper) GetIndexOfStopInStopTimeUpdatesForTrip(tripID, stopID string) int {
	for i, sid := range w.onwardStops[tripID] {
		if sid == stopID {
			return i
		}
	}
	return -1
}

func (w *GTFSRTWrapper) GetStartDateForTrip(tripID string) string  { return w.tripDate[tripID] }
func (w *GTFSRTWrapper) GetOriginTimeForTrip(tripID string) string { return "" }

func (w *GTFSRTWrapper) GetVehiclePositionTimestamp(tripID string) int64 {
	if ts, ok := w.vehicleTS[tripID]; ok {
		return ts
	}
	return 0
}

func (w *GTFSRTWrapper) GetTimestampForTrip(tripID string) int64 {
	return w.GetVehiclePositionTimestamp(tripID)
}

func (w *GTFSRTWrapper) GetTimestampForFeedMessage() int64 { return w.headerTimestamp }

// Vehicle accessors
func (w *GTFSRTWrapper) GetVehicleRefForTrip(tripID string) string { return w.tripVehicleRef[tripID] }

func (w *GTFSRTWrapper) GetVehicleLatForTrip(tripID string) (float64, bool) {
	v, ok := w.tripLat[tripID]
	return v, ok
}

func (w *GTFSRTWrapper) GetVehicleLonForTrip(tripID string) (float64, bool) {
	v, ok := w.tripLon[tripID]
	return v, ok
}

func (w *GTFSRTWrapper) GetVehicleBearingForTrip(tripID string) (float64, bool) {
	v, ok := w.tripBearing[tripID]
	return v, ok
}

// GetVehicleSpeedForTrip returns the speed in m/s from VehiclePosition
func (w *GTFSRTWrapper) GetVehicleSpeedForTrip(tripID string) (float64, bool) {
	v, ok := w.tripSpeed[tripID]
	return v, ok
}

// GetScheduleRelationshipForStop returns the schedule_relationship for a stop (0=SCHEDULED, 1=SKIPPED, 2=NO_DATA)
func (w *GTFSRTWrapper) GetScheduleRelationshipForStop(tripID, stopID string) int32 {
	if m, ok := w.schedRelByStop[tripID]; ok {
		return m[stopID]
	}
	return 0 // Default: SCHEDULED
}

// GetOccupancyStatusForTrip returns the occupancy_status from TripUpdate (0-8, -1 if not available)
func (w *GTFSRTWrapper) GetOccupancyStatusForTrip(tripID string) int32 {
	if status, ok := w.tripOccupancy[tripID]; ok {
		return status
	}
	return -1 // Not available
}

// GetCongestionLevelForTrip returns the congestion_level from VehiclePosition (0-4, -1 if not available)
func (w *GTFSRTWrapper) GetCongestionLevelForTrip(tripID string) int32 {
	if level, ok := w.tripCongestion[tripID]; ok {
		return level
	}
	return -1 // Not available
}

// Alerts placeholders
func (w *GTFSRTWrapper) GetAllTripsWithAlert() []string { return nil }

func (w *GTFSRTWrapper) GetTrainsWithAlertFilterObject(trips []string) map[string]bool {
	return map[string]bool{}
}

func (w *GTFSRTWrapper) GetStopsWithAlertFilterObject(trips []string) map[string]bool {
	return map[string]bool{}
}

func (w *GTFSRTWrapper) GetRoutesWithAlertFilterObject(trips []string) map[string]bool {
	return map[string]bool{}
}

// Accessors for alerts for SX builder
func (w *GTFSRTWrapper) GetAlerts() []RTAlert                        { return w.alerts }
func (w *GTFSRTWrapper) GetAlertIndicesByRoute(routeID string) []int { return w.alertsByRoute[routeID] }
func (w *GTFSRTWrapper) GetAlertIndicesByStop(stopID string) []int   { return w.alertsByStop[stopID] }
func (w *GTFSRTWrapper) GetAlertIndicesByTrip(tripID string) []int   { return w.alertsByTrip[tripID] }

// Internal parsing methods

func (w *GTFSRTWrapper) parseTripUpdatesFeed(fm *gtfsrtpb.FeedMessage) {
	if fm == nil {
		return
	}
	if fm.Header != nil && fm.Header.Timestamp != nil {
		if ts := int64(*fm.Header.Timestamp); ts > w.headerTimestamp {
			w.headerTimestamp = ts
		}
	}
	for _, e := range fm.Entity {
		if e.TripUpdate != nil && e.TripUpdate.Trip != nil && e.TripUpdate.Trip.TripId != nil {
			tripID := *e.TripUpdate.Trip.TripId
			w.trips[tripID] = struct{}{}
			w.tripsFromTU[tripID] = struct{}{}
			if e.TripUpdate.Trip.RouteId != nil {
				w.tripRoute[tripID] = *e.TripUpdate.Trip.RouteId
			}
			if e.TripUpdate.Trip.DirectionId != nil {
				w.tripDir[tripID] = string(rune(*e.TripUpdate.Trip.DirectionId + '0'))
			}
			if e.TripUpdate.Trip.StartDate != nil {
				w.tripDate[tripID] = *e.TripUpdate.Trip.StartDate
			}
			if e.TripUpdate.Vehicle != nil && e.TripUpdate.Vehicle.Id != nil {
				w.tripVehicleRef[tripID] = *e.TripUpdate.Vehicle.Id
			}
			if len(e.TripUpdate.StopTimeUpdate) > 0 {
				w.onwardStops[tripID] = make([]string, 0, len(e.TripUpdate.StopTimeUpdate))
				w.etaByStop[tripID] = map[string]int64{}
				w.etdByStop[tripID] = map[string]int64{}
				w.schedRelByStop[tripID] = map[string]int32{}
				for _, stu := range e.TripUpdate.StopTimeUpdate {
					if stu.StopId == nil {
						continue
					}
					sid := *stu.StopId
					w.onwardStops[tripID] = append(w.onwardStops[tripID], sid)
					if stu.Arrival != nil && stu.Arrival.Time != nil {
						w.etaByStop[tripID][sid] = int64(*stu.Arrival.Time)
					}
					if stu.Departure != nil && stu.Departure.Time != nil {
						w.etdByStop[tripID][sid] = int64(*stu.Departure.Time)
					}
					if stu.ScheduleRelationship != nil {
						w.schedRelByStop[tripID][sid] = int32(*stu.ScheduleRelationship)
					}
				}
			}
		}
	}
}

func (w *GTFSRTWrapper) parseVehiclePositionsFeed(fm *gtfsrtpb.FeedMessage) {
	if fm == nil {
		return
	}
	if fm.Header != nil && fm.Header.Timestamp != nil {
		if ts := int64(*fm.Header.Timestamp); ts > w.headerTimestamp {
			w.headerTimestamp = ts
		}
	}
	for _, e := range fm.Entity {
		if e.Vehicle != nil {
			var tripID string
			if e.Vehicle.Trip != nil && e.Vehicle.Trip.TripId != nil {
				tripID = *e.Vehicle.Trip.TripId
			}
			if tripID != "" {
				w.trips[tripID] = struct{}{}
				w.tripsFromVP[tripID] = struct{}{}
			}
			if e.Vehicle.Vehicle != nil && e.Vehicle.Vehicle.Id != nil && tripID != "" {
				if _, exists := w.tripVehicleRef[tripID]; !exists {
					w.tripVehicleRef[tripID] = *e.Vehicle.Vehicle.Id
				}
			}
			if e.Vehicle.Position != nil && tripID != "" {
				if e.Vehicle.Position.Latitude != nil {
					w.tripLat[tripID] = float64(*e.Vehicle.Position.Latitude)
				}
				if e.Vehicle.Position.Longitude != nil {
					w.tripLon[tripID] = float64(*e.Vehicle.Position.Longitude)
				}
				if e.Vehicle.Position.Bearing != nil {
					w.tripBearing[tripID] = float64(*e.Vehicle.Position.Bearing)
				}
				if e.Vehicle.Position.Speed != nil {
					w.tripSpeed[tripID] = float64(*e.Vehicle.Position.Speed)
				}
			}
			if e.Vehicle.CongestionLevel != nil && tripID != "" {
				w.tripCongestion[tripID] = int32(*e.Vehicle.CongestionLevel)
			}
			if e.Vehicle.OccupancyStatus != nil && tripID != "" {
				w.tripOccupancy[tripID] = int32(*e.Vehicle.OccupancyStatus)
			}
			if e.Vehicle.Timestamp != nil && tripID != "" {
				w.vehicleTS[tripID] = int64(*e.Vehicle.Timestamp)
			}
		}
	}
}

func (w *GTFSRTWrapper) parseServiceAlertsFeed(fm *gtfsrtpb.FeedMessage) {
	if fm == nil {
		return
	}
	if fm.Header != nil && fm.Header.Timestamp != nil {
		if ts := int64(*fm.Header.Timestamp); ts > w.headerTimestamp {
			w.headerTimestamp = ts
		}
	}
	for _, e := range fm.Entity {
		if e.Alert == nil {
			continue
		}
		a := e.Alert
		ra := RTAlert{
			DescriptionByLang: make(map[string]string),
			URLByLang:         make(map[string]string),
		}
		if e.Id != nil {
			ra.ID = *e.Id
		}
		if a.HeaderText != nil {
			ra.Header = translatedStringToText(a.HeaderText)
		}
		if a.DescriptionText != nil {
			ra.Description = translatedStringToText(a.DescriptionText)
			// Extract all language translations
			ra.DescriptionByLang = translatedStringToMap(a.DescriptionText)
		}
		if a.Url != nil {
			// Extract all language URLs
			ra.URLByLang = translatedStringToMap(a.Url)
		}
		if a.Cause != nil {
			ra.Cause = a.Cause.String()
		}
		if a.Effect != nil {
			ra.Effect = a.Effect.String()
		}
		if a.SeverityLevel != nil {
			ra.Severity = a.SeverityLevel.String()
		}
		if len(a.ActivePeriod) > 0 {
			ap := a.ActivePeriod[0]
			if ap.Start != nil {
				ra.Start = int64(*ap.Start)
			}
			if ap.End != nil {
				ra.End = int64(*ap.End)
			}
		}
		for _, ie := range a.InformedEntity {
			if ie.RouteId != nil {
				rid := *ie.RouteId
				ra.RouteIDs = append(ra.RouteIDs, rid)
			}
			if ie.Trip != nil && ie.Trip.TripId != nil {
				tid := *ie.Trip.TripId
				ra.TripIDs = append(ra.TripIDs, tid)
			}
			if ie.StopId != nil {
				sid := *ie.StopId
				ra.StopIDs = append(ra.StopIDs, sid)
			}
		}
		idx := len(w.alerts)
		w.alerts = append(w.alerts, ra)
		for _, rid := range ra.RouteIDs {
			w.alertsByRoute[rid] = append(w.alertsByRoute[rid], idx)
		}
		for _, sid := range ra.StopIDs {
			w.alertsByStop[sid] = append(w.alertsByStop[sid], idx)
		}
		for _, tid := range ra.TripIDs {
			w.alertsByTrip[tid] = append(w.alertsByTrip[tid], idx)
		}
	}
}

// translatedStringToText returns the best-effort text from a TranslatedString
func translatedStringToText(ts *gtfsrtpb.TranslatedString) string {
	if ts == nil || len(ts.Translation) == 0 {
		return ""
	}
	// Prefer entries with no language tag or first entry
	var first string
	for _, tr := range ts.Translation {
		if tr.Text != nil {
			if tr.Language == nil || *tr.Language == "" {
				return *tr.Text
			}
			if first == "" {
				first = *tr.Text
			}
		}
	}
	return first
}

// translatedStringToMap returns all translations as a map of language -> text
func translatedStringToMap(ts *gtfsrtpb.TranslatedString) map[string]string {
	result := make(map[string]string)
	if ts == nil || len(ts.Translation) == 0 {
		return result
	}
	for _, tr := range ts.Translation {
		if tr.Text != nil {
			lang := "unknown"
			if tr.Language != nil && *tr.Language != "" {
				lang = *tr.Language
			}
			result[lang] = *tr.Text
		}
	}
	return result
}
