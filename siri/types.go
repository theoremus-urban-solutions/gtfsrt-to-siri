package siri

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
	ResponseTimestamp          string               `json:"ResponseTimestamp"`
	ProducerRef                string               `json:"ProducerRef,omitempty"`
	VehicleMonitoringDelivery  []VehicleMonitoring  `json:"VehicleMonitoringDelivery"`
	SituationExchangeDelivery  []SituationExchange  `json:"SituationExchangeDelivery"`
	EstimatedTimetableDelivery []EstimatedTimetable `json:"EstimatedTimetableDelivery"`
}
