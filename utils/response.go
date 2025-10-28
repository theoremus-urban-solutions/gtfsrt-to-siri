package utils

import "github.com/theoremus-urban-solutions/transit-types/siri"

// SiriResponse contains all SIRI delivery types
type SiriResponse struct {
	ResponseTimestamp          string                            `json:"ResponseTimestamp"`
	ProducerRef                string                            `json:"ProducerRef,omitempty"`
	VehicleMonitoringDelivery  []siri.VehicleMonitoringDelivery  `json:"VehicleMonitoringDelivery"`
	SituationExchangeDelivery  []siri.SituationExchangeDelivery  `json:"SituationExchangeDelivery"`
	EstimatedTimetableDelivery []siri.EstimatedTimetableDelivery `json:"EstimatedTimetableDelivery"`
}
