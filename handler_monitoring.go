package gtfsrtsiri

import (
	"net/http"
)

func handleVehicleMonitoringJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := map[string]string{}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	gtfs, _ := NewGTFSIndexFromConfig(Config.GTFS)
	if _, err := parseAndValidateVehicleMonitoring(params, gtfs); err != nil {
		w.WriteHeader(400)
		_, _ = w.Write(buildErrorPayload("vehicleMonitoring", "json", err.Error()))
		return
	}
	rt := NewGTFSRTWrapper(Config.GTFSRT.TripUpdatesURL, Config.GTFSRT.VehiclePositionsURL)
	_ = rt.Refresh()
	conv := NewConverter(gtfs, rt, Config)
	cache := NewConverterCache(conv)
	buf, err := cache.GetVehicleMonitoringResponse(params, "json")
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write(buildErrorPayload("vehicleMonitoring", "json", err.Error()))
		return
	}
	_, _ = w.Write(buf)
}

func handleStopMonitoringJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := map[string]string{}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	gtfs, _ := NewGTFSIndexFromConfig(Config.GTFS)
	if _, err := parseAndValidateStopMonitoring(params, gtfs); err != nil {
		w.WriteHeader(400)
		_, _ = w.Write(buildErrorPayload("stopMonitoring", "json", err.Error()))
		return
	}
	rt := NewGTFSRTWrapper(Config.GTFSRT.TripUpdatesURL, Config.GTFSRT.VehiclePositionsURL)
	_ = rt.Refresh()
	conv := NewConverter(gtfs, rt, Config)
	cache := NewConverterCache(conv)
	buf, err := cache.GetStopMonitoringResponse(params, "json")
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write(buildErrorPayload("stopMonitoring", "json", err.Error()))
		return
	}
	_, _ = w.Write(buf)
}

func handleVehicleMonitoringXML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	params := map[string]string{}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	gtfs, _ := NewGTFSIndexFromConfig(Config.GTFS)
	if _, err := parseAndValidateVehicleMonitoring(params, gtfs); err != nil {
		w.WriteHeader(400)
		_, _ = w.Write(buildErrorPayload("vehicleMonitoring", "xml", err.Error()))
		return
	}
	rt := NewGTFSRTWrapper(Config.GTFSRT.TripUpdatesURL, Config.GTFSRT.VehiclePositionsURL)
	_ = rt.Refresh()
	conv := NewConverter(gtfs, rt, Config)
	cache := NewConverterCache(conv)
	buf, err := cache.GetVehicleMonitoringResponse(params, "xml")
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write(buildErrorPayload("vehicleMonitoring", "xml", err.Error()))
		return
	}
	_, _ = w.Write(buf)
}

func handleStopMonitoringXML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	params := map[string]string{}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	gtfs, _ := NewGTFSIndexFromConfig(Config.GTFS)
	if _, err := parseAndValidateStopMonitoring(params, gtfs); err != nil {
		w.WriteHeader(400)
		_, _ = w.Write(buildErrorPayload("stopMonitoring", "xml", err.Error()))
		return
	}
	rt := NewGTFSRTWrapper(Config.GTFSRT.TripUpdatesURL, Config.GTFSRT.VehiclePositionsURL)
	_ = rt.Refresh()
	conv := NewConverter(gtfs, rt, Config)
	cache := NewConverterCache(conv)
	buf, err := cache.GetStopMonitoringResponse(params, "xml")
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write(buildErrorPayload("stopMonitoring", "xml", err.Error()))
		return
	}
	_, _ = w.Write(buf)
}
