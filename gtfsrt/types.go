package gtfsrt

// RTAlert is a simplified representation of a GTFS-RT Alert for SX building
type RTAlert struct {
	ID          string
	Header      string
	Description string
	Cause       string
	Effect      string
	Severity    string
	Start       int64
	End         int64
	RouteIDs    []string
	StopIDs     []string
	TripIDs     []string
}
