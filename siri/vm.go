package siri

// VehicleMonitoring represents the VehicleMonitoring delivery
type VehicleMonitoring struct {
	ResponseTimestamp string                 `json:"ResponseTimestamp"`
	ValidUntil        string                 `json:"ValidUntil"`
	VehicleActivity   []VehicleActivityEntry `json:"VehicleActivity"`
}

// VehicleActivityEntry represents a single vehicle's activity
type VehicleActivityEntry struct {
	RecordedAtTime          string                  `json:"RecordedAtTime"`
	ValidUntilTime          string                  `json:"ValidUntilTime,omitempty"` // SIRI-VM spec: required per activity
	MonitoredVehicleJourney MonitoredVehicleJourney `json:"MonitoredVehicleJourney"`
}

// MonitoredVehicleJourney contains details about a monitored vehicle journey
type MonitoredVehicleJourney struct {
	LineRef                  string         `json:"LineRef"`
	DirectionRef             any            `json:"DirectionRef,omitempty"`
	FramedVehicleJourneyRef  any            `json:"FramedVehicleJourneyRef,omitempty"`
	VehicleMode              string         `json:"VehicleMode,omitempty"` // bus, tram, metro, etc.
	JourneyPatternRef        string         `json:"JourneyPatternRef,omitempty"`
	PublishedLineName        string         `json:"PublishedLineName,omitempty"`
	OperatorRef              string         `json:"OperatorRef,omitempty"`
	OriginRef                string         `json:"OriginRef,omitempty"`
	OriginName               string         `json:"OriginName,omitempty"`
	DestinationRef           string         `json:"DestinationRef,omitempty"`
	DestinationName          string         `json:"DestinationName,omitempty"`
	OriginAimedDepartureTime string         `json:"OriginAimedDepartureTime,omitempty"`
	SituationRef             any            `json:"SituationRef,omitempty"`
	Monitored                bool           `json:"Monitored"`
	DataSource               string         `json:"DataSource"` // SIRI-VM spec: required codespace
	VehicleLocation          any            `json:"VehicleLocation"`
	Bearing                  *float64       `json:"Bearing,omitempty"`
	Velocity                 *int           `json:"Velocity,omitempty"`
	Occupancy                string         `json:"Occupancy,omitempty"`
	Delay                    string         `json:"Delay,omitempty"` // SIRI-VM spec: required (e.g., "PT16S" or "PT0S")
	InCongestion             *bool          `json:"InCongestion,omitempty"`
	VehicleStatus            string         `json:"VehicleStatus,omitempty"`
	ProgressRate             any            `json:"ProgressRate,omitempty"`
	ProgressStatus           any            `json:"ProgressStatus,omitempty"`
	VehicleRef               string         `json:"VehicleRef"`
	MonitoredCall            *MonitoredCall `json:"MonitoredCall,omitempty"` // SIRI-VM spec: current/previous stop only
	IsCompleteStopSequence   bool           `json:"IsCompleteStopSequence"`  // SIRI-VM spec: required, always false
	OnwardCalls              any            `json:"OnwardCalls,omitempty"`   // Keep for backwards compatibility, but should be removed for VM
}

// VehicleLocation represents the geographical location of a vehicle
type VehicleLocation struct {
	Latitude  *float64 `json:"Latitude"`
	Longitude *float64 `json:"Longitude"`
}

// MonitoredCall represents information about the current/previous stop for VM
type MonitoredCall struct {
	StopPointRef          string           `json:"StopPointRef"`
	Order                 *int             `json:"Order,omitempty"` // Stop order/sequence in trip
	StopPointName         string           `json:"StopPointName,omitempty"`
	VehicleAtStop         *bool            `json:"VehicleAtStop,omitempty"`
	VehicleLocationAtStop *VehicleLocation `json:"VehicleLocationAtStop,omitempty"`
	DestinationDisplay    string           `json:"DestinationDisplay,omitempty"`
}

// SiriCall represents a call at a stop
type SiriCall struct {
	Extensions struct {
		Distances struct {
			PresentableDistance    string   `json:"PresentableDistance"`
			DistanceFromCall       *float64 `json:"DistanceFromCall"`
			StopsFromCall          int      `json:"StopsFromCall"`
			CallDistanceAlongRoute float64  `json:"CallDistanceAlongRoute"`
		} `json:"Distances"`
	} `json:"Extensions"`
	ExpectedArrivalTime   string `json:"ExpectedArrivalTime"`
	ExpectedDepartureTime string `json:"ExpectedDepartureTime"`
	StopPointRef          string `json:"StopPointRef"`
	StopPointName         string `json:"StopPointName"`
	VisitNumber           int    `json:"VisitNumber"`
}
