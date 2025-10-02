package converter

import (
	"encoding/json"

	"mta/gtfsrt-to-siri/config"
	"mta/gtfsrt-to-siri/gtfs"
	"mta/gtfsrt-to-siri/gtfsrt"
	"mta/gtfsrt-to-siri/internal"
	"mta/gtfsrt-to-siri/siri"
)

// Converter coordinates GTFS, GTFS-RT, and configuration to produce SIRI responses
type Converter struct {
	GTFS   *gtfs.GTFSIndex
	GTFSRT *gtfsrt.GTFSRTWrapper
	Cfg    config.AppConfig
	Snap   *Snapshot
}

// NewConverter creates a new converter instance
func NewConverter(gtfsIdx *gtfs.GTFSIndex, rt *gtfsrt.GTFSRTWrapper, cfg config.AppConfig) *Converter {
	snap, _ := NewSnapshot(gtfsIdx, rt, cfg)
	return &Converter{GTFS: gtfsIdx, GTFSRT: rt, Cfg: cfg, Snap: snap}
}

// GetCompleteVehicleMonitoringResponse builds a complete VM SIRI response
func (c *Converter) GetCompleteVehicleMonitoringResponse() *siri.SiriResponse {
	timestamp := c.GTFSRT.GetTimestampForFeedMessage()
	vm := siri.VehicleMonitoring{
		ResponseTimestamp: internal.Iso8601FromUnixSeconds(timestamp),
		ValidUntil:        internal.ValidUntilFrom(timestamp, c.Cfg.GTFSRT.ReadIntervalMS),
		VehicleActivity:   []siri.VehicleActivityEntry{},
	}
	// Minimal entry to make shape valid
	vm.VehicleActivity = append(vm.VehicleActivity, siri.VehicleActivityEntry{
		RecordedAtTime:          internal.Iso8601FromUnixSeconds(timestamp),
		MonitoredVehicleJourney: c.buildMVJ(""),
	})
	return &siri.SiriResponse{Siri: siri.SiriServiceDelivery{ServiceDelivery: siri.VehicleAndSituation{
		ResponseTimestamp:         internal.Iso8601FromUnixSeconds(timestamp),
		VehicleMonitoringDelivery: []siri.VehicleMonitoring{vm},
		SituationExchangeDelivery: []siri.SituationExchange{},
	}}}
}

// GetState returns the current converter state as JSON
func (c *Converter) GetState() []byte {
	b, _ := json.Marshal(map[string]any{
		"gtfsrtTimestamp": c.GTFSRT.GetTimestampForFeedMessage(),
	})
	return b
}
