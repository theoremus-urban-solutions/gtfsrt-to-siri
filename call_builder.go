package gtfsrtsiri

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

func (c *Converter) buildCall(tripID, stopID string) SiriCall {
	var call SiriCall
	call.VisitNumber = 1
	return call
}
