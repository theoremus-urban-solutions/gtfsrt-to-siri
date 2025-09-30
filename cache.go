package gtfsrtsiri

import (
	"bytes"
	"errors"
	"sort"
	"strconv"
	"strings"
)

type ConverterCache struct {
	converter     *Converter
	responseCache map[string][]byte
}

func NewConverterCache(c *Converter) *ConverterCache {
	return &ConverterCache{converter: c, responseCache: map[string][]byte{}}
}

func (cc *ConverterCache) memoKey(args ...string) string {
	var b bytes.Buffer
	for i, a := range args {
		if i > 0 {
			b.WriteByte('|')
		}
		b.WriteString(a)
	}
	return b.String()
}

func (cc *ConverterCache) build(res *SiriResponse, format string) []byte {
	rb := newResponseBuilder()
	if format == "xml" {
		return rb.BuildXML(res)
	}
	return rb.BuildJSON(res)
}

// GetSituationExchangeResponse returns a minimal SiriResponse with an empty SituationExchangeDelivery.
// Placeholder until SX builder is implemented.
func (cc *ConverterCache) GetSituationExchangeResponse(format string) ([]byte, error) {
	ts := cc.converter.GTFSRT.GetTimestampForFeedMessage()
	sx := cc.converter.BuildSituationExchange()
	producerRef := cc.converter.Cfg.GTFS.AgencyID
	if producerRef == "" {
		producerRef = "UNKNOWN"
	}
	res := &SiriResponse{Siri: SiriServiceDelivery{ServiceDelivery: VehicleAndSituation{
		ResponseTimestamp:         iso8601FromUnixSeconds(ts),
		ProducerRef:               producerRef,
		VehicleMonitoringDelivery: []VehicleMonitoring{},
		SituationExchangeDelivery: []SituationExchange{sx},
		StopMonitoringDelivery:    []StopMonitoring{},
	}}}
	return cc.build(res, format), nil
}

func (cc *ConverterCache) selectTripsByVM(params map[string]string) []string {
	// Filters: vehicleref, lineref, directionref
	vehRef := strings.ToLower(params["vehicleref"])
	lineRef := strings.ToLower(params["lineref"])
	dirRef := strings.ToLower(params["directionref"])
	trips := cc.converter.GTFSRT.GetAllMonitoredTrips()
	out := make([]string, 0, len(trips))
	for _, t := range trips {
		if vehRef != "" && strings.ToLower(cc.converter.GTFSRT.GetVehicleRefForTrip(t)) != vehRef {
			continue
		}
		if lineRef != "" {
			rid := strings.ToLower(cc.converter.GTFSRT.GetRouteIDForTrip(t))
			// support fully qualified LineRef like AGENCY_route
			if rid != lineRef {
				if strings.Contains(lineRef, "_") {
					parts := strings.SplitN(lineRef, "_", 2)
					if len(parts) == 2 {
						if rid != parts[1] {
							continue
						}
					}
				} else {
					continue
				}
			}
		}
		if dirRef != "" && strings.ToLower(cc.converter.GTFSRT.GetRouteDirectionForTrip(t)) != dirRef {
			continue
		}
		out = append(out, t)
	}
	return out
}

func (cc *ConverterCache) selectTripsByStop(stopID string, params map[string]string) []string {
	stopIDLower := strings.ToLower(stopID)
	lineRef := strings.ToLower(params["lineref"])     // may be AGENCY_route or route
	dirRef := strings.ToLower(params["directionref"]) // "0" or "1"
	trips := cc.converter.GTFSRT.GetAllMonitoredTrips()
	out := make([]string, 0)
	for _, t := range trips {
		// stop inclusion check
		found := false
		idxAtStop := -1
		stops := cc.converter.GTFSRT.GetOnwardStopIDsForTrip(t)
		for i, sid := range stops {
			if strings.ToLower(sid) == stopIDLower {
				found = true
				idxAtStop = i
				break
			}
		}
		if !found {
			continue
		}
		// optional line filter
		if lineRef != "" {
			rid := strings.ToLower(cc.converter.GTFSRT.GetRouteIDForTrip(t))
			if rid != lineRef {
				if parts := strings.SplitN(lineRef, "_", 2); len(parts) == 2 {
					if rid != parts[1] {
						continue
					}
				} else {
					continue
				}
			}
		}
		// optional direction filter
		if dirRef != "" && strings.ToLower(cc.converter.GTFSRT.GetRouteDirectionForTrip(t)) != dirRef {
			continue
		}
		// ensure the stop is not beyond list (already guaranteed), store trip
		_ = idxAtStop
		out = append(out, t)
	}
	return out
}

func (cc *ConverterCache) buildVMResponse(trips []string, includeCalls bool) *SiriResponse {
	ts := cc.converter.GTFSRT.GetTimestampForFeedMessage()
	vm := VehicleMonitoring{
		ResponseTimestamp: iso8601FromUnixSeconds(ts),
		ValidUntil:        validUntilFrom(ts, cc.converter.Cfg.GTFSRT.ReadIntervalMS),
		VehicleActivity:   []VehicleActivityEntry{},
	}
	for _, t := range trips {
		mvj := cc.converter.buildMVJ(t)
		if !includeCalls {
			mvj.OnwardCalls = nil
		}
		vat := VehicleActivityEntry{RecordedAtTime: iso8601FromUnixSeconds(cc.converter.GTFSRT.GetTimestampForTrip(t)), MonitoredVehicleJourney: mvj}
		vm.VehicleActivity = append(vm.VehicleActivity, vat)
	}
	sx := cc.converter.BuildSituationExchange()
	producerRef := cc.converter.Cfg.GTFS.AgencyID
	if producerRef == "" {
		producerRef = "UNKNOWN"
	}
	return &SiriResponse{Siri: SiriServiceDelivery{ServiceDelivery: VehicleAndSituation{
		ResponseTimestamp:         iso8601FromUnixSeconds(ts),
		ProducerRef:               producerRef,
		VehicleMonitoringDelivery: []VehicleMonitoring{vm},
		SituationExchangeDelivery: []SituationExchange{sx},
	}}}
}

func (cc *ConverterCache) buildVMResponseWithCalls(trips []string, includeCalls bool, maxOnward int, stopID string, stopMonitoring bool) *SiriResponse {
	ts := cc.converter.GTFSRT.GetTimestampForFeedMessage()
	vm := VehicleMonitoring{
		ResponseTimestamp: iso8601FromUnixSeconds(ts),
		ValidUntil:        validUntilFrom(ts, cc.converter.Cfg.GTFSRT.ReadIntervalMS),
		VehicleActivity:   []VehicleActivityEntry{},
	}
	for _, t := range trips {
		mvj := cc.converter.buildMVJ(t)
		if includeCalls {
			mvj.OnwardCalls = cc.converter.buildOnwardCalls(t, maxOnward, stopID, stopMonitoring)
		} else {
			mvj.OnwardCalls = nil
		}
		vat := VehicleActivityEntry{RecordedAtTime: iso8601FromUnixSeconds(cc.converter.GTFSRT.GetTimestampForTrip(t)), MonitoredVehicleJourney: mvj}
		vm.VehicleActivity = append(vm.VehicleActivity, vat)
	}
	sx := cc.converter.BuildSituationExchange()
	producerRef := cc.converter.Cfg.GTFS.AgencyID
	if producerRef == "" {
		producerRef = "UNKNOWN"
	}
	return &SiriResponse{Siri: SiriServiceDelivery{ServiceDelivery: VehicleAndSituation{
		ResponseTimestamp:         iso8601FromUnixSeconds(ts),
		ProducerRef:               producerRef,
		VehicleMonitoringDelivery: []VehicleMonitoring{vm},
		SituationExchangeDelivery: []SituationExchange{sx},
	}}}
}

func (cc *ConverterCache) buildSMResponse(stopID string, trips []string, maxOnward int) *SiriResponse {
	ts := cc.converter.GTFSRT.GetTimestampForFeedMessage()
	sm := StopMonitoring{
		ResponseTimestamp:  iso8601FromUnixSeconds(ts),
		MonitoredStopVisit: []MonitoredStopVisit{},
	}
	for _, t := range trips {
		mvj := cc.converter.buildMVJ(t)
		// For SM: build OnwardCalls starting from selected stop
		mvj.OnwardCalls = cc.converter.buildOnwardCalls(t, maxOnward, stopID, true)
		ms := MonitoredStopVisit{
			RecordedAtTime:          iso8601FromUnixSeconds(cc.converter.GTFSRT.GetTimestampForTrip(t)),
			MonitoringRef:           applyFieldMutators(stopID, cc.converter.Cfg.Converter.FieldMutators.StopPointRef),
			MonitoredVehicleJourney: mvj,
			MonitoredCall:           cc.converter.buildCall(t, stopID),
		}
		// Populate MonitoredCall identifiers and name with mutators
		ms.MonitoredCall.StopPointRef = applyFieldMutators(stopID, cc.converter.Cfg.Converter.FieldMutators.StopPointRef)
		ms.MonitoredCall.StopPointName = cc.converter.GTFS.GetStopName(stopID)
		// Fill MonitoredCall timings
		if eta := cc.converter.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(t, stopID); eta > 0 {
			ms.MonitoredCall.ExpectedArrivalTime = iso8601FromUnixSeconds(eta)
		}
		if etd := cc.converter.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(t, stopID); etd > 0 {
			ms.MonitoredCall.ExpectedDepartureTime = iso8601FromUnixSeconds(etd)
		}
		sm.MonitoredStopVisit = append(sm.MonitoredStopVisit, ms)
	}
	sx := cc.converter.BuildSituationExchange()
	producerRef := cc.converter.Cfg.GTFS.AgencyID
	if producerRef == "" {
		producerRef = "UNKNOWN"
	}
	return &SiriResponse{Siri: SiriServiceDelivery{ServiceDelivery: VehicleAndSituation{
		ResponseTimestamp:         iso8601FromUnixSeconds(ts),
		ProducerRef:               producerRef,
		VehicleMonitoringDelivery: []VehicleMonitoring{},
		SituationExchangeDelivery: []SituationExchange{sx},
		StopMonitoringDelivery:    []StopMonitoring{sm},
	}}}
}

func (cc *ConverterCache) GetStopMonitoringResponse(params map[string]string, format string) ([]byte, error) {
	if params == nil {
		return nil, errors.New("params required")
	}
	stopID := params["monitoringref"]
	maxOnward := -1
	if s := params["maximumnumberofcallsonwards"]; s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			maxOnward = v
		}
	}
	key := cc.memoKey("sm", format, strings.ToLower(stopID), strconv.Itoa(maxOnward), strings.ToLower(params["lineref"]), strings.ToLower(params["directionref"]), params["maximumstopvisits"], params["minimumstopvisitsperline"])
	if buf, ok := cc.responseCache[key]; ok {
		return buf, nil
	}
	trips := cc.selectTripsByStop(stopID, params)
	// Apply maximumStopVisits/minimumStopVisitsPerLine parity behavior using ETA order per route
	maxSV := -1
	minPerLine := -1
	if s := params["maximumstopvisits"]; s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			maxSV = v
		}
	}
	if s := params["minimumstopvisitsperline"]; s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			minPerLine = v
		}
	}
	if maxSV >= 0 {
		// Partition trips by route
		byRoute := map[string][]string{}
		for _, t := range trips {
			rid := cc.converter.GTFSRT.GetRouteIDForTrip(t)
			byRoute[rid] = append(byRoute[rid], t)
		}
		// Sort each route's trips by ETA at stopID ascending
		for rid := range byRoute {
			sort.Slice(byRoute[rid], func(i, j int) bool {
				ti := byRoute[rid][i]
				tj := byRoute[rid][j]
				ei := cc.converter.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(ti, stopID)
				ej := cc.converter.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(tj, stopID)
				if ei == 0 {
					ei = cc.converter.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(ti, stopID)
				}
				if ej == 0 {
					ej = cc.converter.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(tj, stopID)
				}
				return ei < ej
			})
		}
		selected := make([]string, 0, len(trips))
		// First, take minPerLine from each route
		if minPerLine > 0 {
			for _, arr := range byRoute {
				k := minPerLine
				if k > len(arr) {
					k = len(arr)
				}
				selected = append(selected, arr[:k]...)
			}
		}
		// If we still have room, fill globally by next earliest ETA
		if len(selected) < maxSV {
			type cand struct {
				t   string
				eta int64
			}
			cands := make([]cand, 0, len(trips))
			for _, arr := range byRoute {
				for _, t := range arr {
					// skip if already selected
					skip := false
					for _, s := range selected {
						if s == t {
							skip = true
							break
						}
					}
					if skip {
						continue
					}
					e := cc.converter.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(t, stopID)
					if e == 0 {
						e = cc.converter.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(t, stopID)
					}
					cands = append(cands, cand{t: t, eta: e})
				}
			}
			sort.Slice(cands, func(i, j int) bool { return cands[i].eta < cands[j].eta })
			for _, c := range cands {
				if len(selected) >= maxSV {
					break
				}
				selected = append(selected, c.t)
			}
		}
		// Cap at maxSV if overfull
		if len(selected) > maxSV {
			selected = selected[:maxSV]
		}
		trips = selected
	}
	res := cc.buildSMResponse(stopID, trips, maxOnward)
	buf := cc.build(res, format)
	cc.responseCache[key] = buf
	return buf, nil
}

func (cc *ConverterCache) GetVehicleMonitoringResponse(params map[string]string, format string) ([]byte, error) {
	if params == nil {
		return nil, errors.New("params required")
	}
	detail := strings.ToLower(params["vehiclemonitoringdetaillevel"])
	includeCalls := (detail == "calls")
	maxOnward := -1
	if s := params["maximumnumberofcallsonwards"]; s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			maxOnward = v
		}
	}
	key := cc.memoKey("vm", format, detail, strconv.Itoa(maxOnward), strings.ToLower(params["lineref"]), strings.ToLower(params["directionref"]))
	if buf, ok := cc.responseCache[key]; ok {
		return buf, nil
	}
	trips := cc.selectTripsByVM(params)
	// Apply MaximumStopVisits/MinimumStopVisitsPerLine like in SM
	maxSV := -1
	minPerLine := -1
	if s := params["maximumstopvisits"]; s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			maxSV = v
		}
	}
	if s := params["minimumstopvisitsperline"]; s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			minPerLine = v
		}
	}
	if maxSV >= 0 {
		// Partition trips by route
		byRoute := map[string][]string{}
		for _, t := range trips {
			rid := cc.converter.GTFSRT.GetRouteIDForTrip(t)
			byRoute[rid] = append(byRoute[rid], t)
		}
		// Sort each route's trips by earliest ETA across any stop (approx: first onward stop)
		for rid := range byRoute {
			sort.Slice(byRoute[rid], func(i, j int) bool {
				ti := byRoute[rid][i]
				tj := byRoute[rid][j]
				// choose first onward stop ETA as proxy
				etaI := int64(1 << 62)
				etaJ := int64(1 << 62)
				if stops := cc.converter.GTFSRT.GetOnwardStopIDsForTrip(ti); len(stops) > 0 {
					if e := cc.converter.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(ti, stops[0]); e > 0 {
						etaI = e
					} else {
						etaI = cc.converter.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(ti, stops[0])
					}
				}
				if stops := cc.converter.GTFSRT.GetOnwardStopIDsForTrip(tj); len(stops) > 0 {
					if e := cc.converter.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(tj, stops[0]); e > 0 {
						etaJ = e
					} else {
						etaJ = cc.converter.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(tj, stops[0])
					}
				}
				return etaI < etaJ
			})
		}
		selected := make([]string, 0, len(trips))
		if minPerLine > 0 {
			for _, arr := range byRoute {
				k := minPerLine
				if k > len(arr) {
					k = len(arr)
				}
				selected = append(selected, arr[:k]...)
			}
		}
		if len(selected) < maxSV {
			type cand struct {
				t   string
				eta int64
			}
			cands := make([]cand, 0, len(trips))
			for _, arr := range byRoute {
				for _, t := range arr {
					skip := false
					for _, s := range selected {
						if s == t {
							skip = true
							break
						}
					}
					if skip {
						continue
					}
					eta := int64(1 << 62)
					if stops := cc.converter.GTFSRT.GetOnwardStopIDsForTrip(t); len(stops) > 0 {
						if e := cc.converter.GTFSRT.GetExpectedArrivalTimeAtStopForTrip(t, stops[0]); e > 0 {
							eta = e
						} else {
							eta = cc.converter.GTFSRT.GetExpectedDepartureTimeAtStopForTrip(t, stops[0])
						}
					}
					cands = append(cands, cand{t: t, eta: eta})
				}
			}
			sort.Slice(cands, func(i, j int) bool { return cands[i].eta < cands[j].eta })
			for _, c := range cands {
				if len(selected) >= maxSV {
					break
				}
				selected = append(selected, c.t)
			}
		}
		if len(selected) > maxSV {
			selected = selected[:maxSV]
		}
		trips = selected
	}
	res := cc.buildVMResponseWithCalls(trips, includeCalls, maxOnward, "", false)
	buf := cc.build(res, format)
	cc.responseCache[key] = buf
	return buf, nil
}
