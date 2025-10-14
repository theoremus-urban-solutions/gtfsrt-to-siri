package siri

import "github.com/theoremus-urban-solutions/transit-types/siri"

// SiriResponse is the top-level SIRI response structure
type SiriResponse struct {
	Siri SiriServiceDelivery `json:"Siri"`
}

// SiriServiceDelivery wraps the ServiceDelivery element
type SiriServiceDelivery struct {
	ServiceDelivery VehicleAndSituation `json:"ServiceDelivery"`
}

// VehicleAndSituation contains all SIRI delivery types
type VehicleAndSituation struct {
	ResponseTimestamp          string                            `json:"ResponseTimestamp"`
	ProducerRef                string                            `json:"ProducerRef,omitempty"`
	VehicleMonitoringDelivery  []VehicleMonitoringDelivery       `json:"VehicleMonitoringDelivery"`
	SituationExchangeDelivery  []SituationExchangeDelivery       `json:"SituationExchangeDelivery"`
	EstimatedTimetableDelivery []siri.EstimatedTimetableDelivery `json:"EstimatedTimetableDelivery"`
}
