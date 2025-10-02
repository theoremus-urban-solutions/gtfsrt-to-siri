package converter

import (
	"encoding/json"

	"mta/gtfsrt-to-siri/config"
	"mta/gtfsrt-to-siri/gtfs"
	"mta/gtfsrt-to-siri/gtfsrt"
	"mta/gtfsrt-to-siri/siri"
	"mta/gtfsrt-to-siri/tracking"
	"mta/gtfsrt-to-siri/utils"
)

// Converter coordinates GTFS, GTFS-RT, and configuration to produce SIRI responses
type Converter struct {
	GTFS   *gtfs.GTFSIndex
	GTFSRT *gtfsrt.GTFSRTWrapper
	Cfg    config.AppConfig
	Snap   *tracking.Snapshot
}

// NewConverter creates a new converter instance
func NewConverter(gtfsIdx *gtfs.GTFSIndex, rt *gtfsrt.GTFSRTWrapper, cfg config.AppConfig) *Converter {
	snap, _ := tracking.NewSnapshot(gtfsIdx, rt, cfg)
	return &Converter{GTFS: gtfsIdx, GTFSRT: rt, Cfg: cfg, Snap: snap}
}

// GetCompleteVehicleMonitoringResponse builds a complete VM SIRI response
func (c *Converter) GetCompleteVehicleMonitoringResponse() *siri.SiriResponse {
	timestamp := c.GTFSRT.GetTimestampForFeedMessage()
	codespace := c.Cfg.GTFS.AgencyID

	vm := siri.VehicleMonitoring{
		ResponseTimestamp: utils.Iso8601FromUnixSeconds(timestamp),
		ValidUntil:        utils.ValidUntilFrom(timestamp, c.Cfg.GTFSRT.ReadIntervalMS),
		VehicleActivity:   []siri.VehicleActivityEntry{},
	}
	// Minimal entry to make shape valid
	vm.VehicleActivity = append(vm.VehicleActivity, siri.VehicleActivityEntry{
		RecordedAtTime:          utils.Iso8601FromUnixSeconds(timestamp),
		MonitoredVehicleJourney: c.buildMVJ(""),
	})

	// Use shared ServiceDelivery builder (note: formatter package would be better but causes circular dependency)
	sd := siri.VehicleAndSituation{
		ResponseTimestamp:         utils.Iso8601FromUnixSeconds(timestamp),
		ProducerRef:               codespace,
		VehicleMonitoringDelivery: []siri.VehicleMonitoring{vm},
		SituationExchangeDelivery: []siri.SituationExchange{},
	}

	return &siri.SiriResponse{Siri: siri.SiriServiceDelivery{ServiceDelivery: sd}}
}

// GetState returns the current converter state as JSON
func (c *Converter) GetState() []byte {
	b, _ := json.Marshal(map[string]any{
		"gtfsrtTimestamp": c.GTFSRT.GetTimestampForFeedMessage(),
	})
	return b
}
