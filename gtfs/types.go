package gtfs

// Waypoint represents a geographical coordinate
type Waypoint struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// StopTime contains schedule information for a stop on a trip
type StopTime struct {
	ArrivalTime   string
	DepartureTime string
	PickupType    int8
	DropOffType   int8
}
