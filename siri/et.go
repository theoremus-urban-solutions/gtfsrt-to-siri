package siri

// EstimatedTimetableDelivery represents a delivery of estimated timetable data
// According to SIRI-ET specification v1.1 (Entur Nordic Profile)
type EstimatedTimetableDelivery struct {
	Version                      string                         `json:"version" xml:"version,attr"`
	ResponseTimestamp            string                         `json:"ResponseTimestamp" xml:"ResponseTimestamp"`
	EstimatedJourneyVersionFrame []EstimatedJourneyVersionFrame `json:"EstimatedJourneyVersionFrame" xml:"EstimatedJourneyVersionFrame"`
}

// EstimatedJourneyVersionFrame contains a frame of estimated journeys with a timestamp
type EstimatedJourneyVersionFrame struct {
	RecordedAtTime          string                    `json:"RecordedAtTime" xml:"RecordedAtTime"`
	EstimatedVehicleJourney []EstimatedVehicleJourney `json:"EstimatedVehicleJourney" xml:"EstimatedVehicleJourney"`
}

// EstimatedVehicleJourney represents a single journey with estimated times
// Contains continuously updated timetable data with changes for the current operating day
type EstimatedVehicleJourney struct {
	RecordedAtTime          string                  `json:"RecordedAtTime" xml:"RecordedAtTime"`
	LineRef                 string                  `json:"LineRef" xml:"LineRef"`
	DirectionRef            string                  `json:"DirectionRef" xml:"DirectionRef"`
	FramedVehicleJourneyRef FramedVehicleJourneyRef `json:"FramedVehicleJourneyRef" xml:"FramedVehicleJourneyRef"`
	VehicleRef              string                  `json:"VehicleRef,omitempty" xml:"VehicleRef,omitempty"`
	VehicleMode             string                  `json:"VehicleMode,omitempty" xml:"VehicleMode,omitempty"`
	OriginName              string                  `json:"OriginName,omitempty" xml:"OriginName,omitempty"`
	DestinationName         string                  `json:"DestinationName,omitempty" xml:"DestinationName,omitempty"`
	Monitored               bool                    `json:"Monitored" xml:"Monitored"`
	DataSource              string                  `json:"DataSource,omitempty" xml:"DataSource,omitempty"`
	OperatorRef             string                  `json:"OperatorRef,omitempty" xml:"OperatorRef,omitempty"`
	RecordedCalls           []RecordedCall          `json:"RecordedCalls,omitempty" xml:"RecordedCalls>RecordedCall,omitempty"`
	EstimatedCalls          []EstimatedCall         `json:"EstimatedCalls,omitempty" xml:"EstimatedCalls>EstimatedCall,omitempty"`
	IsCompleteStopSequence  bool                    `json:"IsCompleteStopSequence" xml:"IsCompleteStopSequence"`
}

// FramedVehicleJourneyRef uniquely identifies a vehicle journey within a data frame
type FramedVehicleJourneyRef struct {
	DataFrameRef           string `json:"DataFrameRef" xml:"DataFrameRef"`
	DatedVehicleJourneyRef string `json:"DatedVehicleJourneyRef" xml:"DatedVehicleJourneyRef"`
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
