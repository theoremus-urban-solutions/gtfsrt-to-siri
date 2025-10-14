package siri

import transitTypes "github.com/theoremus-urban-solutions/transit-types/siri"

// SituationExchangeDelivery represents the SIRI-SX delivery structure
// Based on SIRI-SX specification v1.1 (Entur Nordic Profile)
// Spec: https://enturas.atlassian.net/wiki/spaces/PUBLIC/pages/637370605/SIRI-SX
type SituationExchangeDelivery struct {
	Version           string               `json:"version" xml:"version,attr"`
	ResponseTimestamp string               `json:"ResponseTimestamp" xml:"ResponseTimestamp"`
	Situations        []PtSituationElement `json:"Situations" xml:"Situations>PtSituationElement"`
}

// PtSituationElement represents a single public transport situation (alert/disruption)
// Cardinality: 1:* per SituationExchangeDelivery
type PtSituationElement struct {
	CreationTime      string                  `json:"CreationTime" xml:"CreationTime"`
	ParticipantRef    string                  `json:"ParticipantRef" xml:"ParticipantRef"`
	SituationNumber   string                  `json:"SituationNumber" xml:"SituationNumber"`
	Version           *int                    `json:"Version,omitempty" xml:"Version,omitempty"`
	Source            *SituationSource        `json:"Source,omitempty" xml:"Source,omitempty"`
	VersionedAtTime   string                  `json:"VersionedAtTime,omitempty" xml:"VersionedAtTime,omitempty"`
	Progress          string                  `json:"Progress" xml:"Progress"` // open|closed
	ValidityPeriod    []ValidityPeriod        `json:"ValidityPeriod" xml:"ValidityPeriod"`
	UndefinedReason   string                  `json:"UndefinedReason,omitempty" xml:"UndefinedReason,omitempty"` // Always empty per spec
	Severity          string                  `json:"Severity,omitempty" xml:"Severity,omitempty"`
	Priority          *int                    `json:"Priority,omitempty" xml:"Priority,omitempty"`
	ReportType        string                  `json:"ReportType" xml:"ReportType"` // general|incident
	Planned           *bool                   `json:"Planned,omitempty" xml:"Planned,omitempty"`
	Keywords          []string                `json:"Keywords,omitempty" xml:"Keywords>Keyword,omitempty"`
	Summary           []NaturalLanguageString `json:"Summary,omitempty" xml:"Summary,omitempty"`
	Description       []NaturalLanguageString `json:"Description,omitempty" xml:"Description,omitempty"`
	Detail            []NaturalLanguageString `json:"Detail,omitempty" xml:"Detail,omitempty"`
	Advice            []NaturalLanguageString `json:"Advice,omitempty" xml:"Advice,omitempty"`
	Internal          []NaturalLanguageString `json:"Internal,omitempty" xml:"Internal,omitempty"`
	Affects           *Affects                `json:"Affects,omitempty" xml:"Affects,omitempty"`
	Consequences      *Consequences           `json:"Consequences,omitempty" xml:"Consequences,omitempty"`
	PublishingActions *PublishingActions      `json:"PublishingActions,omitempty" xml:"PublishingActions,omitempty"`
	InfoLinks         []InfoLink              `json:"InfoLinks,omitempty" xml:"InfoLinks>InfoLink,omitempty"`
}

// SituationSource represents the source of the situation message
type SituationSource struct {
	SourceType string `json:"SourceType,omitempty" xml:"SourceType,omitempty"`
}

// ValidityPeriod represents a time period with start and optional end time
type ValidityPeriod struct {
	StartTime string `json:"StartTime" xml:"StartTime"`
	EndTime   string `json:"EndTime,omitempty" xml:"EndTime,omitempty"`
}

// NaturalLanguageString represents text with a language attribute
type NaturalLanguageString struct {
	Lang string `json:"lang,omitempty" xml:"lang,attr,omitempty"`
	Text string `json:"text" xml:",chardata"`
}

// InfoLink represents a URL with optional language attribute
type InfoLink struct {
	Uri   string                  `json:"Uri" xml:"Uri"`
	Label []NaturalLanguageString `json:"Label,omitempty" xml:"Label,omitempty"`
}

// Affects represents the scope of the situation
type Affects struct {
	Networks        *AffectedNetworks        `json:"Networks,omitempty" xml:"Networks,omitempty"`
	StopPoints      *AffectedStopPoints      `json:"StopPoints,omitempty" xml:"StopPoints,omitempty"`
	StopPlaces      *AffectedStopPlaces      `json:"StopPlaces,omitempty" xml:"StopPlaces,omitempty"`
	VehicleJourneys *AffectedVehicleJourneys `json:"VehicleJourneys,omitempty" xml:"VehicleJourneys,omitempty"`
}

// AffectedNetworks represents affected networks
type AffectedNetworks struct {
	AffectedNetwork []AffectedNetwork `json:"AffectedNetwork" xml:"AffectedNetwork"`
}

// AffectedNetwork represents an affected network
type AffectedNetwork struct {
	NetworkRef    string                  `json:"NetworkRef,omitempty" xml:"NetworkRef,omitempty"`
	NetworkName   []NaturalLanguageString `json:"NetworkName,omitempty" xml:"NetworkName,omitempty"`
	AffectedLines *AffectedLines          `json:"AffectedLine,omitempty" xml:"AffectedLine,omitempty"`
}

// AffectedLines represents affected lines
type AffectedLines struct {
	AffectedLine []AffectedLine `json:"AffectedLine" xml:"AffectedLine"`
}

// AffectedLine represents an affected line/route
type AffectedLine struct {
	LineRef  string                  `json:"LineRef" xml:"LineRef"`
	LineName []NaturalLanguageString `json:"LineName,omitempty" xml:"LineName,omitempty"`
	Routes   *AffectedRoutes         `json:"Routes,omitempty" xml:"Routes,omitempty"`
	Sections *AffectedSections       `json:"Sections,omitempty" xml:"Sections,omitempty"`
}

// AffectedRoutes represents affected routes
type AffectedRoutes struct {
	AffectedRoute []AffectedRoute `json:"AffectedRoute" xml:"AffectedRoute"`
}

// AffectedRoute represents an affected route
type AffectedRoute struct {
	RouteRef     string              `json:"RouteRef,omitempty" xml:"RouteRef,omitempty"`
	DirectionRef string              `json:"DirectionRef,omitempty" xml:"DirectionRef,omitempty"`
	StopPoints   *AffectedStopPoints `json:"StopPoints,omitempty" xml:"StopPoints,omitempty"`
	Sections     *AffectedSections   `json:"Sections,omitempty" xml:"Sections,omitempty"`
}

// AffectedSections represents affected sections
type AffectedSections struct {
	AffectedSection []AffectedSection `json:"AffectedSection" xml:"AffectedSection"`
}

// AffectedSection represents an affected section
type AffectedSection struct {
	IndirectSectionRef *IndirectSectionRef `json:"IndirectSectionRef,omitempty" xml:"IndirectSectionRef,omitempty"`
}

// IndirectSectionRef represents a section defined by first and last quay
type IndirectSectionRef struct {
	FirstQuayRef string `json:"FirstQuayRef" xml:"FirstQuayRef"`
	LastQuayRef  string `json:"LastQuayRef" xml:"LastQuayRef"`
}

// AffectedStopPoints represents affected stop points
type AffectedStopPoints struct {
	AffectedStopPoint []AffectedStopPoint `json:"AffectedStopPoint" xml:"AffectedStopPoint"`
}

// AffectedStopPoint represents an affected stop
type AffectedStopPoint struct {
	StopPointRef  string                  `json:"StopPointRef" xml:"StopPointRef"`
	StopPointName []NaturalLanguageString `json:"StopPointName,omitempty" xml:"StopPointName,omitempty"`
	StopCondition []string                `json:"StopCondition,omitempty" xml:"StopCondition,omitempty"`
}

// AffectedStopPlaces represents affected stop places
type AffectedStopPlaces struct {
	AffectedStopPlace []AffectedStopPlace `json:"AffectedStopPlace" xml:"AffectedStopPlace"`
}

// AffectedStopPlace represents an affected stop place
type AffectedStopPlace struct {
	StopPlaceRef            string                   `json:"StopPlaceRef" xml:"StopPlaceRef"`
	PlaceName               []NaturalLanguageString  `json:"PlaceName,omitempty" xml:"PlaceName,omitempty"`
	AccessibilityAssessment *AccessibilityAssessment `json:"AccessibilityAssessment,omitempty" xml:"AccessibilityAssessment,omitempty"`
	AffectedComponents      *AffectedComponents      `json:"AffectedComponents,omitempty" xml:"AffectedComponents,omitempty"`
}

// AccessibilityAssessment represents accessibility information
type AccessibilityAssessment struct {
	MobilityImpairedAccess bool                     `json:"MobilityImpairedAccess" xml:"MobilityImpairedAccess"`
	Limitations            *AccessibilityLimitation `json:"Limitations,omitempty" xml:"Limitations,omitempty"`
}

// AccessibilityLimitation represents accessibility limitations
type AccessibilityLimitation struct {
	WheelchairAccess    string `json:"WheelchairAccess" xml:"WheelchairAccess"`
	StepFreeAccess      string `json:"StepFreeAccess" xml:"StepFreeAccess"`
	EscalatorFreeAccess string `json:"EscalatorFreeAccess" xml:"EscalatorFreeAccess"`
	LiftFreeAccess      string `json:"LiftFreeAccess" xml:"LiftFreeAccess"`
}

// AffectedComponents represents affected components
type AffectedComponents struct {
	AffectedComponent []AffectedComponent `json:"AffectedComponent" xml:"AffectedComponent"`
}

// AffectedComponent represents an affected component
type AffectedComponent struct {
	ComponentRef      string `json:"ComponentRef,omitempty" xml:"ComponentRef,omitempty"`
	ComponentType     string `json:"ComponentType" xml:"ComponentType"`
	AccessFeatureType string `json:"AccessFeatureType,omitempty" xml:"AccessFeatureType,omitempty"`
}

// AffectedVehicleJourneys represents affected vehicle journeys
type AffectedVehicleJourneys struct {
	AffectedVehicleJourney []AffectedVehicleJourney `json:"AffectedVehicleJourney" xml:"AffectedVehicleJourney"`
}

// AffectedVehicleJourney represents an affected vehicle journey
type AffectedVehicleJourney struct {
	VehicleJourneyRef        string                                `json:"VehicleJourneyRef,omitempty" xml:"VehicleJourneyRef,omitempty"`
	DatedVehicleJourneyRef   string                                `json:"DatedVehicleJourneyRef,omitempty" xml:"DatedVehicleJourneyRef,omitempty"`
	FramedVehicleJourneyRef  *transitTypes.FramedVehicleJourneyRef `json:"FramedVehicleJourneyRef,omitempty" xml:"FramedVehicleJourneyRef,omitempty"`
	Operator                 *AffectedOperator                     `json:"Operator,omitempty" xml:"Operator,omitempty"`
	LineRef                  string                                `json:"LineRef,omitempty" xml:"LineRef,omitempty"`
	Route                    []AffectedRoute                       `json:"Route,omitempty" xml:"Route,omitempty"`
	OriginAimedDepartureTime string                                `json:"OriginAimedDepartureTime,omitempty" xml:"OriginAimedDepartureTime,omitempty"`
}

// AffectedOperator represents an affected operator
type AffectedOperator struct {
	OperatorRef  string                  `json:"OperatorRef" xml:"OperatorRef"`
	OperatorName []NaturalLanguageString `json:"OperatorName,omitempty" xml:"OperatorName,omitempty"`
}

// Consequences represents the consequences of a situation
type Consequences struct {
	Consequence []Consequence `json:"Consequence" xml:"Consequence"`
}

// Consequence represents a single consequence
type Consequence struct {
	Condition  string      `json:"Condition,omitempty" xml:"Condition,omitempty"`
	Severity   string      `json:"Severity,omitempty" xml:"Severity,omitempty"`
	Affects    *Affects    `json:"Affects,omitempty" xml:"Affects,omitempty"`
	Advice     *Advice     `json:"Advice,omitempty" xml:"Advice,omitempty"`
	Blocking   *Blocking   `json:"Blocking,omitempty" xml:"Blocking,omitempty"`
	Boarding   *Boarding   `json:"Boarding,omitempty" xml:"Boarding,omitempty"`
	Delays     *Delays     `json:"Delays,omitempty" xml:"Delays,omitempty"`
	Casualties *Casualties `json:"Casualties,omitempty" xml:"Casualties,omitempty"`
}

// Advice represents advice for passengers
type Advice struct {
	Details []NaturalLanguageString `json:"Details,omitempty" xml:"Details,omitempty"`
}

// Blocking represents blocking information
type Blocking struct {
	JourneyPlanner bool `json:"JourneyPlanner,omitempty" xml:"JourneyPlanner,omitempty"`
	RealTime       bool `json:"RealTime,omitempty" xml:"RealTime,omitempty"`
}

// Boarding represents boarding information
type Boarding struct {
	ArrivalBoardingActivity   string `json:"ArrivalBoardingActivity,omitempty" xml:"ArrivalBoardingActivity,omitempty"`
	DepartureBoardingActivity string `json:"DepartureBoardingActivity,omitempty" xml:"DepartureBoardingActivity,omitempty"`
}

// Delays represents delay information
type Delays struct {
	Delay string `json:"Delay,omitempty" xml:"Delay,omitempty"`
}

// Casualties represents casualty information
type Casualties struct {
	NumberOfDeaths  *int `json:"NumberOfDeaths,omitempty" xml:"NumberOfDeaths,omitempty"`
	NumberOfInjured *int `json:"NumberOfInjured,omitempty" xml:"NumberOfInjured,omitempty"`
}

// PublishingActions represents publishing actions
type PublishingActions struct {
	PublishToWebAction     *PublishToWebAction     `json:"PublishToWebAction,omitempty" xml:"PublishToWebAction,omitempty"`
	PublishToMobileAction  *PublishToMobileAction  `json:"PublishToMobileAction,omitempty" xml:"PublishToMobileAction,omitempty"`
	PublishToDisplayAction *PublishToDisplayAction `json:"PublishToDisplayAction,omitempty" xml:"PublishToDisplayAction,omitempty"`
	PublishToAlertsAction  *PublishToAlertsAction  `json:"PublishToAlertsAction,omitempty" xml:"PublishToAlertsAction,omitempty"`
	ManualAction           *ManualAction           `json:"ManualAction,omitempty" xml:"ManualAction,omitempty"`
	NotifyByEmailAction    *NotifyByEmailAction    `json:"NotifyByEmailAction,omitempty" xml:"NotifyByEmailAction,omitempty"`
	NotifyBySmsAction      *NotifyBySmsAction      `json:"NotifyBySmsAction,omitempty" xml:"NotifyBySmsAction,omitempty"`
	NotifyUserAction       *NotifyUserAction       `json:"NotifyUserAction,omitempty" xml:"NotifyUserAction,omitempty"`
}

// PublishToWebAction represents a publish to web action
type PublishToWebAction struct {
	Incidents bool `json:"Incidents,omitempty" xml:"Incidents,omitempty"`
	HomePage  bool `json:"HomePage,omitempty" xml:"HomePage,omitempty"`
	Ticker    bool `json:"Ticker,omitempty" xml:"Ticker,omitempty"`
}

// PublishToMobileAction represents a publish to mobile action
type PublishToMobileAction struct {
	Incidents bool `json:"Incidents,omitempty" xml:"Incidents,omitempty"`
	HomePage  bool `json:"HomePage,omitempty" xml:"HomePage,omitempty"`
}

// PublishToDisplayAction represents a publish to display action
type PublishToDisplayAction struct {
	OnPlace bool `json:"OnPlace,omitempty" xml:"OnPlace,omitempty"`
	OnBoard bool `json:"OnBoard,omitempty" xml:"OnBoard,omitempty"`
}

// PublishToAlertsAction represents a publish to alerts action
type PublishToAlertsAction struct {
	ByEmail bool `json:"ByEmail,omitempty" xml:"ByEmail,omitempty"`
	BySms   bool `json:"BySms,omitempty" xml:"BySms,omitempty"`
}

// ManualAction represents a manual action
type ManualAction struct {
	Description []NaturalLanguageString `json:"Description,omitempty" xml:"Description,omitempty"`
}

// NotifyByEmailAction represents a notify by email action
type NotifyByEmailAction struct {
	Email string `json:"email,omitempty" xml:"email,omitempty"`
}

// NotifyBySmsAction represents a notify by SMS action
type NotifyBySmsAction struct {
	Phone string `json:"Phone,omitempty" xml:"Phone,omitempty"`
}

// NotifyUserAction represents a notify user action
type NotifyUserAction struct {
	WorkgroupRef []string `json:"WorkgroupRef,omitempty" xml:"WorkgroupRef,omitempty"`
	UserRef      []string `json:"UserRef,omitempty" xml:"UserRef,omitempty"`
}

// SituationExchange is kept for backwards compatibility
type SituationExchange struct {
	Situations any `json:"Situations" xml:"Situations"`
}
