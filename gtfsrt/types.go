package gtfsrt

// RTAlert is a simplified representation of a GTFS-RT Alert for SX building
type RTAlert struct {
	ID                string
	Header            string
	Description       string
	DescriptionByLang map[string]string // language -> description text
	URLByLang         map[string]string // language -> URL
	Cause             string
	Effect            string
	Severity          string
	Start             int64
	End               int64
	RouteIDs          []string
	StopIDs           []string
	TripIDs           []string
}
