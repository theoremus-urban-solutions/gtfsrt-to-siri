package converter

import (
	"encoding/json"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/config"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfs"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/tracking"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
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

	// Get all monitored trips and build MVJ for each
	trips := c.GTFSRT.GetAllMonitoredTrips()
	for _, tripID := range trips {
		mvj := c.buildMVJ(tripID)
		tripTimestamp := c.GTFSRT.GetTimestampForTrip(tripID)
		entry := siri.VehicleActivityEntry{
			RecordedAtTime:          utils.Iso8601FromUnixSeconds(tripTimestamp),
			ValidUntilTime:          utils.ValidUntilFrom(tripTimestamp, c.Cfg.GTFSRT.ReadIntervalMS),
			MonitoredVehicleJourney: mvj,
		}
		vm.VehicleActivity = append(vm.VehicleActivity, entry)
	}

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
