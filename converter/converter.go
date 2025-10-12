package converter

import (
	"encoding/json"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfs"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/tracking"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

// Converter coordinates GTFS, GTFS-RT, and options to produce SIRI responses.
// This converter is data-source agnostic and config-free.
type Converter struct {
	gtfs   *gtfs.GTFSIndex
	gtfsrt *gtfsrt.GTFSRTWrapper
	opts   ConverterOptions
	snap   *tracking.Snapshot
}

// NewConverter creates a new converter instance.
//
// Example:
//
//	gtfs, _ := gtfs.NewGTFSIndexFromBytes(gtfsBytes, "AGENCY")
//	rt, _ := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, saBytes)
//	opts := converter.ConverterOptions{
//	    AgencyID:       "AGENCY",
//	    ReadIntervalMS: 30000,
//	    FieldMutators:  converter.FieldMutators{},
//	}
//	conv := converter.NewConverter(gtfs, rt, opts)
func NewConverter(gtfsIdx *gtfs.GTFSIndex, rt *gtfsrt.GTFSRTWrapper, opts ConverterOptions) *Converter {
	snap := tracking.NewSnapshot(gtfsIdx, rt, opts.AgencyID)
	return &Converter{
		gtfs:   gtfsIdx,
		gtfsrt: rt,
		opts:   opts,
		snap:   snap,
	}
}

// GetCompleteVehicleMonitoringResponse builds a complete VM SIRI response
func (c *Converter) GetCompleteVehicleMonitoringResponse() *siri.SiriResponse {
	timestamp := c.gtfsrt.GetTimestampForFeedMessage()
	codespace := c.opts.AgencyID

	vm := siri.VehicleMonitoring{
		ResponseTimestamp: utils.Iso8601FromUnixSeconds(timestamp),
		ValidUntil:        utils.ValidUntilFrom(timestamp, int(c.opts.ReadIntervalMS)),
		VehicleActivity:   []siri.VehicleActivityEntry{},
	}

	// Get trips from VehiclePositions only (VM should only include trips with position data)
	trips := c.gtfsrt.GetTripsFromVehiclePositions()
	for _, tripID := range trips {
		mvj := c.buildMVJ(tripID)
		tripTimestamp := c.gtfsrt.GetTimestampForTrip(tripID)
		entry := siri.VehicleActivityEntry{
			RecordedAtTime:          utils.Iso8601FromUnixSeconds(tripTimestamp),
			ValidUntilTime:          utils.ValidUntilFrom(tripTimestamp, int(c.opts.ReadIntervalMS)),
			MonitoredVehicleJourney: mvj,
		}
		vm.VehicleActivity = append(vm.VehicleActivity, entry)
	}

	// Use shared ServiceDelivery builder
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
		"gtfsrtTimestamp": c.gtfsrt.GetTimestampForFeedMessage(),
	})
	return b
}
