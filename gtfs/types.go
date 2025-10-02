package gtfs

// Waypoint represents a geographical coordinate
type Waypoint struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}
