package gtfsrtsiri

type SituationExchange struct {
	Situations any `json:"Situations"`
}

// SX structures (minimal subset)
type PtSituationElement struct {
	SituationNumber   string `json:"SituationNumber"`
	Summary           string `json:"Summary"`
	Description       string `json:"Description"`
	Severity          string `json:"Severity"`
	Cause             string `json:"Cause"`
	Effect            string `json:"Effect"`
	PublicationWindow struct {
		StartTime string `json:"StartTime"`
		EndTime   string `json:"EndTime"`
	} `json:"PublicationWindow"`
	Affects struct {
		VehicleJourneys []AffectedVehicleJourney `json:"VehicleJourneys"`
		Routes          []AffectedRoute          `json:"Routes"`
		StopPoints      []AffectedStopPoint      `json:"StopPoints"`
	} `json:"Affects"`
	Consequences []Consequence `json:"Consequences,omitempty"`
}

type AffectedVehicleJourney struct {
	DatedVehicleJourneyRef string `json:"DatedVehicleJourneyRef"`
	LineRef                string `json:"LineRef,omitempty"`
	DirectionRef           string `json:"DirectionRef,omitempty"`
}

type AffectedRoute struct {
	LineRef string `json:"LineRef"`
}

type AffectedStopPoint struct {
	StopPointRef string `json:"StopPointRef"`
}

type Consequence struct {
	Condition string `json:"Condition"`
}

// BuildSituationExchange converts parsed GTFS-RT alerts to a simple SX delivery
func (c *Converter) BuildSituationExchange() SituationExchange {
	alerts := c.GTFSRT.GetAlerts()
	elements := make([]PtSituationElement, 0, len(alerts))
	for _, a := range alerts {
		el := PtSituationElement{
			SituationNumber: a.ID,
			Summary:         a.Header,
			Description:     a.Description,
			Severity:        mapGTFSRTSeverityToSIRI(a.Severity),
			Cause:           mapGTFSRTCauseToSIRI(a.Cause),
			Effect:          mapGTFSRTEffectToSIRI(a.Effect),
		}
		if a.Start > 0 {
			el.PublicationWindow.StartTime = iso8601FromUnixSeconds(a.Start)
		}
		if a.End > 0 {
			el.PublicationWindow.EndTime = iso8601FromUnixSeconds(a.End)
		}
		// Build VehicleJourneys with LineRef and DirectionRef (dedupe by trip)
		seenTrips := map[string]bool{}
		for _, tid := range a.TripIDs {
			if seenTrips[tid] {
				continue
			}
			seenTrips[tid] = true
			vj := AffectedVehicleJourney{
				DatedVehicleJourneyRef: TripKeyForConverter(tid, c.Cfg.GTFS.AgencyID, c.GTFSRT.GetStartDateForTrip(tid)),
			}
			// LineRef from route (optionally agency-prefixed)
			if rid := c.GTFSRT.GetRouteIDForTrip(tid); rid != "" {
				lineRef := rid
				if c.Cfg.GTFS.AgencyID != "" {
					lineRef = c.Cfg.GTFS.AgencyID + "_" + rid
				}
				vj.LineRef = lineRef
			}
			if dir := c.GTFSRT.GetRouteDirectionForTrip(tid); dir != "" {
				vj.DirectionRef = dir
			}
			el.Affects.VehicleJourneys = append(el.Affects.VehicleJourneys, vj)
		}
		for _, rid := range a.RouteIDs {
			lineRef := rid
			if c.Cfg.GTFS.AgencyID != "" {
				lineRef = c.Cfg.GTFS.AgencyID + "_" + rid
			}
			el.Affects.Routes = append(el.Affects.Routes, AffectedRoute{LineRef: lineRef})
		}
		for _, sid := range a.StopIDs {
			el.Affects.StopPoints = append(el.Affects.StopPoints, AffectedStopPoint{StopPointRef: applyFieldMutators(sid, c.Cfg.Converter.FieldMutators.StopPointRef)})
		}
		// Add Consequences derived from normalized Effect
		if cond := effectToCondition(el.Effect); cond != "" {
			el.Consequences = []Consequence{{Condition: cond}}
		}
		elements = append(elements, el)
	}
	return SituationExchange{Situations: elements}
}

// Minimal normalizers for GTFS-RT → SIRI values
func mapGTFSRTSeverityToSIRI(gtfsrtSeverity string) string {
	switch gtfsrtSeverity {
	case "UNKNOWN_SEVERITY":
		return "unknown"
	case "NO_IMPACT", "OTHER_SEVERITY", "INFO":
		return "normal"
	case "WARNING":
		return "warning"
	case "SEVERE":
		return "severe"
	case "VERY_SEVERE":
		return "verySevere"
	default:
		return "unknown"
	}
}

func mapGTFSRTCauseToSIRI(gtfsrtCause string) string {
	switch gtfsrtCause {
	case "TECHNICAL_PROBLEM":
		return "TechnicalProblem"
	case "CONSTRUCTION":
		return "Construction"
	case "MAINTENANCE":
		return "Maintenance"
	case "WEATHER":
		return "Weather"
	case "ACCIDENT":
		return "Accident"
	case "MEDICAL_EMERGENCY":
		return "MedicalEmergency"
	case "POLICE_ACTIVITY":
		return "PoliceActivity"
	case "STRIKE":
		return "IndustrialAction"
	case "DEMONSTRATION":
		return "Demonstration"
	case "OTHER_CAUSE":
		return "Other"
	default:
		return "Unknown"
	}
}

func mapGTFSRTEffectToSIRI(gtfsrtEffect string) string {
	switch gtfsrtEffect {
	case "NO_SERVICE":
		return "NoService"
	case "REDUCED_SERVICE":
		return "ReducedService"
	case "SIGNIFICANT_DELAYS":
		return "SevereDelays"
	case "DETOUR":
		return "Detour"
	case "ADDITIONAL_SERVICE":
		return "AdditionalService"
	case "MODIFIED_SERVICE":
		return "AlteredService"
	case "STOP_MOVED":
		return "StopPointRelocation"
	case "STOPS_CHANGED":
		return "StopPointChange"
	case "ACCESSIBILITY_ISSUE":
		return "AccessibilityIssue"
	case "MAINTENANCE_WORK":
		return "MaintenanceWork"
	case "CONSTRUCTION":
		return "Construction"
	case "DELAY":
		return "Delay"
	case "CANCELLATION":
		return "Cancellation"
	default:
		return "Other"
	}
}

// Map SIRI Effect → SIRI Consequence Condition (minimal set)
func effectToCondition(effect string) string {
	switch effect {
	case "NoService":
		return "NoService"
	case "ReducedService":
		return "ReducedService"
	case "SevereDelays":
		return "SevereDelays"
	case "Delay":
		return "Delayed"
	case "Detour":
		return "Diversion"
	case "Cancellation":
		return "Cancellation"
	default:
		return ""
	}
}
