package siri

// VehicleMonitoring represents the SIRI-VM delivery structure
// Based on SIRI-VM specification v1.1 (Entur Nordic Profile)
// Spec: https://enturas.atlassian.net/wiki/spaces/PUBLIC/pages/637370425/SIRI-VM
type VehicleMonitoring struct {
	Version           string            `json:"version" xml:"version,attr"`
	ResponseTimestamp string            `json:"ResponseTimestamp" xml:"ResponseTimestamp"`
	VehicleActivity   []VehicleActivity `json:"VehicleActivity" xml:"VehicleActivity"`
}

// VehicleActivity represents a single vehicle activity entry
// Cardinality: 1:* per VehicleMonitoring
type VehicleActivity struct {
	RecordedAtTime          string                   `json:"RecordedAtTime" xml:"RecordedAtTime"`
	ValidUntilTime          string                   `json:"ValidUntilTime,omitempty" xml:"ValidUntilTime,omitempty"`
	ProgressBetweenStops    *ProgressBetweenStops    `json:"ProgressBetweenStops,omitempty" xml:"ProgressBetweenStops,omitempty"`
	MonitoredVehicleJourney *MonitoredVehicleJourney `json:"MonitoredVehicleJourney" xml:"MonitoredVehicleJourney"`
}

// ProgressBetweenStops represents progress along the current ServiceLink
type ProgressBetweenStops struct {
	LinkDistance float64 `json:"LinkDistance,omitempty" xml:"LinkDistance,omitempty"`
	Percentage   float64 `json:"Percentage" xml:"Percentage"`
}

// MonitoredVehicleJourney represents a real-time monitored vehicle journey
type MonitoredVehicleJourney struct {
	LineRef                 string                   `json:"LineRef" xml:"LineRef"`
	DirectionRef            string                   `json:"DirectionRef,omitempty" xml:"DirectionRef,omitempty"`
	FramedVehicleJourneyRef *FramedVehicleJourneyRef `json:"FramedVehicleJourneyRef,omitempty" xml:"FramedVehicleJourneyRef,omitempty"`
	VehicleMode             string                   `json:"VehicleMode,omitempty" xml:"VehicleMode,omitempty"`
	OperatorRef             string                   `json:"OperatorRef,omitempty" xml:"OperatorRef,omitempty"`
	OriginRef               string                   `json:"OriginRef,omitempty" xml:"OriginRef,omitempty"`
	OriginName              string                   `json:"OriginName,omitempty" xml:"OriginName,omitempty"`
	DestinationRef          string                   `json:"DestinationRef,omitempty" xml:"DestinationRef,omitempty"`
	DestinationName         string                   `json:"DestinationName,omitempty" xml:"DestinationName,omitempty"`
	Monitored               *bool                    `json:"Monitored,omitempty" xml:"Monitored,omitempty"`
	DataSource              string                   `json:"DataSource,omitempty" xml:"DataSource,omitempty"`
	VehicleLocation         *Location                `json:"VehicleLocation,omitempty" xml:"VehicleLocation,omitempty"`
	Bearing                 *float64                 `json:"Bearing,omitempty" xml:"Bearing,omitempty"`
	Velocity                *int                     `json:"Velocity,omitempty" xml:"Velocity,omitempty"`
	Occupancy               string                   `json:"Occupancy,omitempty" xml:"Occupancy,omitempty"`
	Delay                   string                   `json:"Delay,omitempty" xml:"Delay,omitempty"`
	InCongestion            *bool                    `json:"InCongestion,omitempty" xml:"InCongestion,omitempty"`
	VehicleStatus           string                   `json:"VehicleStatus,omitempty" xml:"VehicleStatus,omitempty"`
	VehicleJourneyRef       string                   `json:"VehicleJourneyRef,omitempty" xml:"VehicleJourneyRef,omitempty"`
	VehicleRef              string                   `json:"VehicleRef" xml:"VehicleRef"`
	MonitoredCall           *MonitoredCall           `json:"MonitoredCall,omitempty" xml:"MonitoredCall,omitempty"`
	IsCompleteStopSequence  bool                     `json:"IsCompleteStopSequence" xml:"IsCompleteStopSequence"`
}

// FramedVehicleJourneyRef represents a reference to a vehicle journey with date
type FramedVehicleJourneyRef struct {
	DataFrameRef           string `json:"DataFrameRef" xml:"DataFrameRef"`
	DatedVehicleJourneyRef string `json:"DatedVehicleJourneyRef" xml:"DatedVehicleJourneyRef"`
}

// Location represents a geospatial point
type Location struct {
	Longitude float64 `json:"Longitude" xml:"Longitude"`
	Latitude  float64 `json:"Latitude" xml:"Latitude"`
}

// MonitoredCall represents information about the current/previous stop
type MonitoredCall struct {
	StopPointRef          string    `json:"StopPointRef" xml:"StopPointRef"`
	Order                 *int      `json:"Order,omitempty" xml:"Order,omitempty"`
	StopPointName         string    `json:"StopPointName,omitempty" xml:"StopPointName,omitempty"`
	VehicleAtStop         *bool     `json:"VehicleAtStop,omitempty" xml:"VehicleAtStop,omitempty"`
	VehicleLocationAtStop *Location `json:"VehicleLocationAtStop,omitempty" xml:"VehicleLocationAtStop,omitempty"`
	DestinationDisplay    string    `json:"DestinationDisplay,omitempty" xml:"DestinationDisplay,omitempty"`
}
