package siri

// EstimatedTimetable delivery types
type EstimatedTimetable struct {
	ResponseTimestamp            string                         `json:"ResponseTimestamp"`
	EstimatedJourneyVersionFrame []EstimatedJourneyVersionFrame `json:"EstimatedJourneyVersionFrame"`
}

// EstimatedJourneyVersionFrame contains a frame of estimated journeys
type EstimatedJourneyVersionFrame struct {
	RecordedAtTime          string                    `json:"RecordedAtTime"`
	EstimatedVehicleJourney []EstimatedVehicleJourney `json:"EstimatedVehicleJourney"`
}

// EstimatedVehicleJourney represents a single journey with estimated times
type EstimatedVehicleJourney struct {
	RecordedAtTime          string                  `json:"RecordedAtTime"`
	LineRef                 string                  `json:"LineRef"`
	VehicleRef              string                  `json:"VehicleRef,omitempty"`
	DirectionRef            string                  `json:"DirectionRef"`
	FramedVehicleJourneyRef FramedVehicleJourneyRef `json:"FramedVehicleJourneyRef"`
	VehicleMode             string                  `json:"VehicleMode,omitempty"`
	OriginName              string                  `json:"OriginName,omitempty"`
	DestinationName         string                  `json:"DestinationName,omitempty"`
	Monitored               bool                    `json:"Monitored"`
	DataSource              string                  `json:"DataSource,omitempty"`
	OperatorRef             string                  `json:"OperatorRef,omitempty"`
	RecordedCalls           []RecordedCall          `json:"RecordedCalls,omitempty"`
	EstimatedCalls          []EstimatedCall         `json:"EstimatedCalls,omitempty"`
	IsCompleteStopSequence  bool                    `json:"IsCompleteStopSequence"`
}

// FramedVehicleJourneyRef uniquely identifies a vehicle journey
type FramedVehicleJourneyRef struct {
	DataFrameRef           string `json:"DataFrameRef"`
	DatedVehicleJourneyRef string `json:"DatedVehicleJourneyRef"`
}

// RecordedCall represents a stop that has already been visited
type RecordedCall struct {
	StopPointRef        string `json:"StopPointRef" xml:"StopPointRef"`
	Order               int    `json:"Order" xml:"Order"`
	StopPointName       string `json:"StopPointName,omitempty" xml:"StopPointName,omitempty"`
	Cancellation        bool   `json:"Cancellation,omitempty" xml:"Cancellation"`
	RequestStop         bool   `json:"RequestStop,omitempty" xml:"RequestStop"`
	AimedArrivalTime    string `json:"AimedArrivalTime,omitempty" xml:"AimedArrivalTime,omitempty"`
	ActualArrivalTime   string `json:"ActualArrivalTime,omitempty" xml:"ActualArrivalTime,omitempty"`
	AimedDepartureTime  string `json:"AimedDepartureTime,omitempty" xml:"AimedDepartureTime,omitempty"`
	ActualDepartureTime string `json:"ActualDepartureTime,omitempty" xml:"ActualDepartureTime,omitempty"`
}

// EstimatedCall represents a stop that has not yet been visited
type EstimatedCall struct {
	StopPointRef          string `json:"StopPointRef" xml:"StopPointRef"`
	Order                 int    `json:"Order" xml:"Order"`
	StopPointName         string `json:"StopPointName,omitempty" xml:"StopPointName,omitempty"`
	Cancellation          bool   `json:"Cancellation,omitempty" xml:"Cancellation"`
	RequestStop           bool   `json:"RequestStop,omitempty" xml:"RequestStop"`
	AimedArrivalTime      string `json:"AimedArrivalTime,omitempty" xml:"AimedArrivalTime,omitempty"`
	ExpectedArrivalTime   string `json:"ExpectedArrivalTime,omitempty" xml:"ExpectedArrivalTime,omitempty"`
	AimedDepartureTime    string `json:"AimedDepartureTime,omitempty" xml:"AimedDepartureTime,omitempty"`
	ExpectedDepartureTime string `json:"ExpectedDepartureTime,omitempty" xml:"ExpectedDepartureTime,omitempty"`
	ArrivalStatus         string `json:"ArrivalStatus,omitempty" xml:"ArrivalStatus,omitempty"`
	DepartureStatus       string `json:"DepartureStatus,omitempty" xml:"DepartureStatus,omitempty"`
}
