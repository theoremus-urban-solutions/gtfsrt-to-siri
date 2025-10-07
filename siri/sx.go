package siri

// SituationExchange represents the SituationExchange delivery
type SituationExchange struct {
	Situations any `json:"Situations"`
}

// LocalizedText represents text with a language attribute
type LocalizedText struct {
	Lang string `xml:"lang,attr" json:"lang"`
	Text string `xml:",chardata" json:"text"`
}

// InfoLink represents a URL with a language attribute
type InfoLink struct {
	Lang string `xml:"lang,attr" json:"lang"`
	URL  string `xml:",chardata" json:"url"`
}

// PtSituationElement represents a single public transport situation (alert/disruption)
type PtSituationElement struct {
	ParticipantRef    string `json:"ParticipantRef,omitempty"`
	SituationNumber   string `json:"SituationNumber"`
	SourceType        string `json:"SourceType,omitempty"`
	Progress          string `json:"Progress,omitempty"`
	PublicationWindow struct {
		StartTime string `json:"StartTime"`
		EndTime   string `json:"EndTime"`
	} `json:"PublicationWindow"`
	Severity     string          `json:"Severity"`
	ReportType   string          `json:"ReportType,omitempty"`
	Summaries    []LocalizedText `xml:"Summary" json:"Summary,omitempty"`
	Descriptions []LocalizedText `xml:"Description" json:"Description,omitempty"`
	Affects      struct {
		Networks        []AffectedNetwork        `json:"Networks,omitempty"`
		VehicleJourneys []AffectedVehicleJourney `json:"VehicleJourneys,omitempty"`
		StopPoints      []AffectedStopPoint      `json:"StopPoints,omitempty"`
	} `json:"Affects"`
	InfoLinks    []InfoLink    `xml:"InfoLinks>InfoLink" json:"InfoLinks,omitempty"`
	Consequences []Consequence `json:"Consequences,omitempty"`
}

// AffectedNetwork represents an affected network
type AffectedNetwork struct {
	NetworkRef    string         `json:"NetworkRef,omitempty"`
	AffectedLines []AffectedLine `json:"AffectedLines,omitempty"`
}

// AffectedLine represents an affected line/route
type AffectedLine struct {
	LineRef        string          `json:"LineRef"`
	AffectedRoutes []AffectedRoute `json:"AffectedRoutes,omitempty"`
}

// AffectedRoute represents an affected route
type AffectedRoute struct {
	DirectionRef string              `json:"DirectionRef,omitempty"`
	StopPoints   []AffectedStopPoint `json:"StopPoints,omitempty"`
}

// AffectedVehicleJourney represents an affected vehicle journey
type AffectedVehicleJourney struct {
	DatedVehicleJourneyRef string `json:"DatedVehicleJourneyRef"`
	LineRef                string `json:"LineRef,omitempty"`
	DirectionRef           string `json:"DirectionRef,omitempty"`
}

// AffectedStopPoint represents an affected stop
type AffectedStopPoint struct {
	StopPointRef string `json:"StopPointRef"`
}

// Consequence represents the consequence of a situation
type Consequence struct {
	Condition string `json:"Condition"`
}
