package gtfsrtsiri

import (
	"io"
	"net/http"
	"time"

	gtfsrtpb "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

type GTFSRTWrapper struct {
	tripUpdatesURL      string
	vehiclePositionsURL string

	trips           map[string]struct{}
	vehicleTS       map[string]int64
	headerTimestamp int64

	tripRoute   map[string]string           // trip_id -> route_id
	tripDir     map[string]string           // trip_id -> direction (string)
	tripDate    map[string]string           // trip_id -> start_date (YYYYMMDD)
	onwardStops map[string][]string         // trip_id -> ordered stop_ids
	etaByStop   map[string]map[string]int64 // trip_id -> stop_id -> arrival epoch
	etdByStop   map[string]map[string]int64 // trip_id -> stop_id -> departure epoch

	tripVehicleRef map[string]string  // trip_id -> vehicle id
	tripLat        map[string]float64 // trip_id -> lat
	tripLon        map[string]float64 // trip_id -> lon
	tripBearing    map[string]float64 // trip_id -> bearing
}

func NewGTFSRTWrapper(tripUpdatesURL, vehiclePositionsURL string) *GTFSRTWrapper {
	return &GTFSRTWrapper{
		tripUpdatesURL:      tripUpdatesURL,
		vehiclePositionsURL: vehiclePositionsURL,
		trips:               map[string]struct{}{},
		vehicleTS:           map[string]int64{},
		headerTimestamp:     time.Now().Unix(),
		tripRoute:           map[string]string{},
		tripDir:             map[string]string{},
		tripDate:            map[string]string{},
		onwardStops:         map[string][]string{},
		etaByStop:           map[string]map[string]int64{},
		etdByStop:           map[string]map[string]int64{},
		tripVehicleRef:      map[string]string{},
		tripLat:             map[string]float64{},
		tripLon:             map[string]float64{},
		tripBearing:         map[string]float64{},
	}
}

func (w *GTFSRTWrapper) Refresh() error {
	w.trips = map[string]struct{}{}
	w.vehicleTS = map[string]int64{}
	w.tripRoute = map[string]string{}
	w.tripDir = map[string]string{}
	w.tripDate = map[string]string{}
	w.onwardStops = map[string][]string{}
	w.etaByStop = map[string]map[string]int64{}
	w.etdByStop = map[string]map[string]int64{}
	w.tripVehicleRef = map[string]string{}
	w.tripLat = map[string]float64{}
	w.tripLon = map[string]float64{}
	w.tripBearing = map[string]float64{}
	w.headerTimestamp = 0
	if w.tripUpdatesURL != "" {
		if fm, err := fetchFeed(w.tripUpdatesURL); err == nil && fm != nil {
			if fm.Header != nil && fm.Header.Timestamp != nil {
				if ts := int64(*fm.Header.Timestamp); ts > w.headerTimestamp {
					w.headerTimestamp = ts
				}
			}
			for _, e := range fm.Entity {
				if e.TripUpdate != nil && e.TripUpdate.Trip != nil && e.TripUpdate.Trip.TripId != nil {
					tripID := *e.TripUpdate.Trip.TripId
					w.trips[tripID] = struct{}{}
					if e.TripUpdate.Trip.RouteId != nil {
						w.tripRoute[tripID] = *e.TripUpdate.Trip.RouteId
					}
					if e.TripUpdate.Trip.DirectionId != nil {
						w.tripDir[tripID] = string(rune(*e.TripUpdate.Trip.DirectionId + '0'))
					}
					if e.TripUpdate.Trip.StartDate != nil {
						w.tripDate[tripID] = *e.TripUpdate.Trip.StartDate
					}
					if len(e.TripUpdate.StopTimeUpdate) > 0 {
						w.onwardStops[tripID] = make([]string, 0, len(e.TripUpdate.StopTimeUpdate))
						w.etaByStop[tripID] = map[string]int64{}
						w.etdByStop[tripID] = map[string]int64{}
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
						}
					}
				}
			}
		}
	}
	if w.vehiclePositionsURL != "" {
		if fm, err := fetchFeed(w.vehiclePositionsURL); err == nil && fm != nil {
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
					}
					if e.Vehicle.Vehicle != nil && e.Vehicle.Vehicle.Id != nil && tripID != "" {
						w.tripVehicleRef[tripID] = *e.Vehicle.Vehicle.Id
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
					}
					if e.Vehicle.Timestamp != nil && tripID != "" {
						w.vehicleTS[tripID] = int64(*e.Vehicle.Timestamp)
					}
				}
			}
		}
	}
	if w.headerTimestamp == 0 {
		w.headerTimestamp = time.Now().Unix()
	}
	return nil
}

func fetchFeed(url string) (*gtfsrtpb.FeedMessage, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var fm gtfsrtpb.FeedMessage
	if err := proto.Unmarshal(b, &fm); err != nil {
		return nil, err
	}
	return &fm, nil
}

func (w *GTFSRTWrapper) GetAllMonitoredTrips() []string {
	ids := make([]string, 0, len(w.trips))
	for id := range w.trips {
		ids = append(ids, id)
	}
	return ids
}

func (w *GTFSRTWrapper) GetGTFSTripKeyForRealtimeTripKey(tripID string) string { return tripID }

// TripKeyForConverter returns tripKey according to Config.Converter.TripKeyStrategy
func TripKeyForConverter(tripID, agency, startDate string) string {
	s := Config.Converter.TripKeyStrategy
	switch s {
	case "startDateTrip":
		if startDate != "" {
			return startDate + "_" + tripID
		}
	case "agencyTrip":
		if agency != "" {
			return agency + "_" + tripID
		}
	case "agencyStartDateTrip":
		key := tripID
		if startDate != "" {
			key = startDate + "_" + key
		}
		if agency != "" {
			key = agency + "_" + key
		}
		return key
	}
	return tripID
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

// New vehicle accessors
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
