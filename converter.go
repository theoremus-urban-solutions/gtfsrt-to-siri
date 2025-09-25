package gtfsrtsiri

import (
	"encoding/json"
)

type SiriResponse struct {
	Siri SiriServiceDelivery `json:"Siri"`
}

type SiriServiceDelivery struct {
	ServiceDelivery VehicleAndSituation `json:"ServiceDelivery"`
}

type VehicleAndSituation struct {
	ResponseTimestamp         string              `json:"ResponseTimestamp"`
	VehicleMonitoringDelivery []VehicleMonitoring `json:"VehicleMonitoringDelivery"`
	SituationExchangeDelivery []SituationExchange `json:"SituationExchangeDelivery"`
	StopMonitoringDelivery    []StopMonitoring    `json:"StopMonitoringDelivery"`
}

type VehicleMonitoring struct {
	ResponseTimestamp string                 `json:"ResponseTimestamp"`
	ValidUntil        string                 `json:"ValidUntil"`
	VehicleActivity   []VehicleActivityEntry `json:"VehicleActivity"`
}

type VehicleActivityEntry struct {
	RecordedAtTime          string                  `json:"RecordedAtTime"`
	MonitoredVehicleJourney MonitoredVehicleJourney `json:"MonitoredVehicleJourney"`
}

// StopMonitoring delivery types
type StopMonitoring struct {
	ResponseTimestamp  string               `json:"ResponseTimestamp"`
	MonitoredStopVisit []MonitoredStopVisit `json:"MonitoredStopVisit"`
}

type MonitoredStopVisit struct {
	RecordedAtTime          string                  `json:"RecordedAtTime"`
	MonitoringRef           string                  `json:"MonitoringRef"`
	MonitoredVehicleJourney MonitoredVehicleJourney `json:"MonitoredVehicleJourney"`
	MonitoredCall           SiriCall                `json:"MonitoredCall"`
}

type Converter struct {
	GTFS   *GTFSIndex
	GTFSRT *GTFSRTWrapper
	Cfg    AppConfig
	Snap   *Snapshot
}

func NewConverter(gtfs *GTFSIndex, rt *GTFSRTWrapper, cfg AppConfig) *Converter {
	snap, _ := NewSnapshot(gtfs, rt, cfg)
	return &Converter{GTFS: gtfs, GTFSRT: rt, Cfg: cfg, Snap: snap}
}

func (c *Converter) GetCompleteVehicleMonitoringResponse() *SiriResponse {
	timestamp := c.GTFSRT.GetTimestampForFeedMessage()
	vm := VehicleMonitoring{
		ResponseTimestamp: iso8601FromUnixSeconds(timestamp),
		ValidUntil:        validUntilFrom(timestamp, c.Cfg.GTFSRT.ReadIntervalMS),
		VehicleActivity:   []VehicleActivityEntry{},
	}
	// Minimal entry to make shape valid
	vm.VehicleActivity = append(vm.VehicleActivity, VehicleActivityEntry{
		RecordedAtTime:          iso8601FromUnixSeconds(timestamp),
		MonitoredVehicleJourney: c.buildMVJ(""),
	})
	return &SiriResponse{Siri: SiriServiceDelivery{ServiceDelivery: VehicleAndSituation{
		ResponseTimestamp:         iso8601FromUnixSeconds(timestamp),
		VehicleMonitoringDelivery: []VehicleMonitoring{vm},
		SituationExchangeDelivery: []SituationExchange{},
	}}}
}

func (c *Converter) GetState() []byte {
	b, _ := json.Marshal(map[string]any{
		"gtfsrtTimestamp": c.GTFSRT.GetTimestampForFeedMessage(),
	})
	return b
}
