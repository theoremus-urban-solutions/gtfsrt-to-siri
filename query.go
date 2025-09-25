package gtfsrtsiri

import (
	"encoding/json"
	"strconv"
	"strings"
)

type QueryError struct{ Msg string }

func (e *QueryError) Error() string { return e.Msg }

// QueryProcessor parity helpers
func normalizeDetailLevel(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" || s == "normal" {
		return "normal", nil
	}
	if s == "calls" {
		return "calls", nil
	}
	return "", &QueryError{Msg: "Unsupported VehicleMonitoringDetailLevel: " + s}
}

func parseNonNegativeInt(s string) (int, error) {
	if s == "" {
		return -1, nil
	}
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || v < 0 {
		return -1, &QueryError{Msg: "Numeric parameter must be a non-negative integer."}
	}
	return v, nil
}

func ensureRouteExists(lineRef string, gtfs *GTFSIndex) (string, error) {
	if lineRef == "" {
		return "", nil
	}
	lr := strings.TrimSpace(lineRef)
	// accept AGENCY_route or bare route
	rid := lr
	if strings.Contains(lr, "_") {
		parts := strings.SplitN(lr, "_", 2)
		if len(parts) == 2 {
			rid = parts[1]
		}
	}
	if _, ok := gtfs.routes[rid]; ok {
		return rid, nil
	}
	if _, ok := gtfs.routeShortNames[rid]; ok {
		return rid, nil
	}
	return "", &QueryError{Msg: "No such route: " + lineRef}
}

func ensureStopExists(stopID string, gtfs *GTFSIndex) error {
	if stopID == "" {
		return &QueryError{Msg: "You must provide a MonitoringRef."}
	}
	if _, ok := gtfs.stopNames[stopID]; !ok {
		return &QueryError{Msg: "No such stop: " + stopID + "."}
	}
	return nil
}

func ensureOperatorValid(op string, gtfs *GTFSIndex) error {
	if op == "" {
		return nil
	}
	for _, ag := range gtfs.GetAllAgencyIDs() {
		if strings.EqualFold(op, ag) {
			return nil
		}
	}
	return &QueryError{Msg: "No such operator: " + op}
}

func parseAndValidateStopMonitoring(params map[string]string, gtfs *GTFSIndex) (map[string]string, error) {
	m := map[string]string{}
	for k, v := range params {
		m[lower(k)] = strings.TrimSpace(v)
		m["_"+k] = v
	}
	if err := ensureStopExists(m["monitoringref"], gtfs); err != nil {
		return nil, err
	}
	if dr := m["directionref"]; dr != "" && dr != "0" && dr != "1" {
		return nil, &QueryError{Msg: "DirectionRef must be either 0 or 1."}
	}
	if _, err := ensureRouteExists(m["lineref"], gtfs); err != nil {
		return nil, err
	}
	if err := ensureOperatorValid(m["operatorref"], gtfs); err != nil {
		return nil, err
	}
	if _, err := parseNonNegativeInt(m["maximumstopvisits"]); err != nil {
		return nil, err
	}
	if _, err := parseNonNegativeInt(m["minimumstopvisitsperline"]); err != nil {
		return nil, err
	}
	if _, err := parseNonNegativeInt(m["maximumnumberofcallsonwards"]); err != nil {
		return nil, err
	}
	return m, nil
}

func parseAndValidateVehicleMonitoring(params map[string]string, gtfs *GTFSIndex) (map[string]string, error) {
	m := map[string]string{}
	for k, v := range params {
		m[lower(k)] = strings.TrimSpace(v)
		m["_"+k] = v
	}
	if dr := m["directionref"]; dr != "" && dr != "0" && dr != "1" {
		return nil, &QueryError{Msg: "DirectionRef must be either 0 or 1."}
	}
	if _, err := ensureRouteExists(m["lineref"], gtfs); err != nil {
		return nil, err
	}
	if err := ensureOperatorValid(m["operatorref"], gtfs); err != nil {
		return nil, err
	}
	if _, err := parseNonNegativeInt(m["maximumnumberofcallsonwards"]); err != nil {
		return nil, err
	}
	if dl, err := normalizeDetailLevel(m["vehiclemonitoringdetaillevel"]); err != nil {
		return nil, err
	} else {
		m["vehiclemonitoringdetaillevel"] = dl
	}
	return m, nil
}

func lower(s string) string {
	bs := []rune(s)
	for i, r := range bs {
		if r >= 'A' && r <= 'Z' {
			bs[i] = r + 32
		}
	}
	return string(bs)
}

func buildErrorPayload(callType, format, msg string) []byte {
	type siriErr struct {
		Siri struct {
			ServiceDelivery struct {
				ErrorCondition struct {
					Description string `json:"Description"`
				} `json:"ErrorCondition"`
			} `json:"ServiceDelivery"`
		} `json:"Siri"`
	}
	var e siriErr
	e.Siri.ServiceDelivery.ErrorCondition.Description = msg
	b, _ := json.Marshal(e)
	return b
}
