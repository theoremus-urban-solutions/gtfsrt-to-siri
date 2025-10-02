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
	MonitoredVehicleJourney MonitoredVehicleJourney `json:"MonitoredVehicleJourney"`
}

// MonitoredVehicleJourney contains details about a monitored vehicle journey
type MonitoredVehicleJourney struct {
	LineRef                  string   `json:"LineRef"`
	DirectionRef             any      `json:"DirectionRef"`
	FramedVehicleJourneyRef  any      `json:"FramedVehicleJourneyRef"`
	JourneyPatternRef        string   `json:"JourneyPatternRef"`
	PublishedLineName        string   `json:"PublishedLineName"`
	OperatorRef              string   `json:"OperatorRef"`
	OriginRef                string   `json:"OriginRef"`
	DestinationRef           string   `json:"DestinationRef"`
	DestinationName          string   `json:"DestinationName"`
	OriginAimedDepartureTime string   `json:"OriginAimedDepartureTime"`
	SituationRef             any      `json:"SituationRef"`
	Monitored                bool     `json:"Monitored"`
	VehicleLocation          any      `json:"VehicleLocation"`
	Bearing                  *float64 `json:"Bearing"`
	ProgressRate             any      `json:"ProgressRate"`
	ProgressStatus           any      `json:"ProgressStatus"`
	VehicleRef               string   `json:"VehicleRef"`
	OnwardCalls              any      `json:"OnwardCalls"`
}

// VehicleLocation represents the geographical location of a vehicle
type VehicleLocation struct {
	Latitude  *float64 `json:"Latitude"`
	Longitude *float64 `json:"Longitude"`
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
