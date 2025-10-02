package converter

import (
	"mta/gtfsrt-to-siri/gtfsrt"
	"mta/gtfsrt-to-siri/utils"
	"mta/gtfsrt-to-siri/siri"
)

func (c *Converter) BuildSituationExchange() siri.SituationExchange {
	alerts := c.GTFSRT.GetAlerts()
	elements := make([]siri.PtSituationElement, 0, len(alerts))
	now := c.GTFSRT.GetTimestampForFeedMessage()
	for _, a := range alerts {
		severity, effectPrefix := mapGTFSRTEffectToSIRISeverity(a.Effect)
		description := a.Description
		if effectPrefix != "" {
			description = effectPrefix + ": " + a.Description
		}
		// Build situation number with codespace prefix
		codespace := c.Cfg.GTFS.AgencyID
		if codespace == "" {
			codespace = "UNKNOWN"
		}
		situationNumber := codespace + ":SituationNumber:" + a.ID

		el := siri.PtSituationElement{
			ParticipantRef:  codespace,
			SituationNumber: situationNumber,
			SourceType:      "directReport",
			Severity:        severity,
			ReportType:      mapGTFSRTCauseToReportType(a.Cause),
			Summary:         mapGTFSRTCauseToSummary(a.Cause),
			Description:     description,
		}
		if a.Start > 0 {
			el.PublicationWindow.StartTime = utils.Iso8601FromUnixSeconds(a.Start)
		}
		if a.End > 0 {
			el.PublicationWindow.EndTime = utils.Iso8601FromUnixSeconds(a.End)
		}
		// Set Progress based on validity period
		if a.End > 0 && a.End < now {
			el.Progress = "closed"
		} else {
			el.Progress = "open"
		}
		// Build Affects structure based on GTFS-RT informed_entity
		// According to affects.md mapping:
		// - Route-only alerts -> Networks > siri.AffectedLine
		// - Trip alerts -> VehicleJourneys
		// - Stop-only alerts -> StopPoints at Affects level

		// Build VehicleJourneys for trip-level alerts
		seenTrips := map[string]bool{}
		for _, tid := range a.TripIDs {
			if seenTrips[tid] {
				continue
			}
			seenTrips[tid] = true
			vj := siri.AffectedVehicleJourney{
				DatedVehicleJourneyRef: gtfsrt.TripKeyForConverter(tid, c.Cfg.GTFS.AgencyID, c.GTFSRT.GetStartDateForTrip(tid)),
			}
			// LineRef with codespace prefix
			if rid := c.GTFSRT.GetRouteIDForTrip(tid); rid != "" {
				vj.LineRef = codespace + ":Line:" + rid
			}
			if dir := c.GTFSRT.GetRouteDirectionForTrip(tid); dir != "" {
				vj.DirectionRef = dir
			}
			el.Affects.VehicleJourneys = append(el.Affects.VehicleJourneys, vj)
		}

		// Build Networks > siri.AffectedLine for route-level alerts
		if len(a.RouteIDs) > 0 {
			network := siri.AffectedNetwork{
				NetworkRef: codespace + ":Network:" + codespace,
			}
			for _, rid := range a.RouteIDs {
				affectedLine := siri.AffectedLine{
					LineRef: codespace + ":Line:" + rid,
				}
				network.AffectedLines = append(network.AffectedLines, affectedLine)
			}
			el.Affects.Networks = append(el.Affects.Networks, network)
		}

		// Build StopPoints for stop-only alerts (at Affects level)
		for _, sid := range a.StopIDs {
			el.Affects.StopPoints = append(el.Affects.StopPoints, siri.AffectedStopPoint{
				StopPointRef: applyFieldMutators(sid, c.Cfg.Converter.FieldMutators.StopPointRef),
			})
		}
		// Add siri.Consequences derived from GTFS-RT Effect
		if cond := effectToCondition(a.Effect); cond != "" {
			el.Consequences = []siri.Consequence{{Condition: cond}}
		}
		elements = append(elements, el)
	}
	return siri.SituationExchange{Situations: elements}
}

// mapGTFSRTEffectToSIRISeverity maps GTFS-RT Effect to SIRI Severity
// Returns (severity, effectPrefix) where effectPrefix is prepended to Description if not empty
func mapGTFSRTEffectToSIRISeverity(gtfsrtEffect string) (string, string) {
	switch gtfsrtEffect {
	case "NO_SERVICE":
		return "noService", ""
	case "REDUCED_SERVICE":
		return "severe", ""
	case "SIGNIFICANT_DELAYS":
		return "severe", ""
	case "DETOUR":
		return "slight", ""
	case "ADDITIONAL_SERVICE":
		return "normal", ""
	case "MODIFIED_SERVICE":
		return "slight", "Modified Service"
	case "OTHER_EFFECT":
		return "undefined", "Other"
	case "UNKNOWN_EFFECT":
		return "undefined", ""
	case "STOP_MOVED":
		return "slight", ""
	case "NO_EFFECT":
		return "noImpact", ""
	case "ACCESSIBILITY_ISSUE":
		return "undefined", "Accessibility Issue"
	default:
		return "undefined", ""
	}
}

func mapGTFSRTCauseToSummary(gtfsrtCause string) string {
	switch gtfsrtCause {
	case "UNKNOWN_CAUSE":
		return "Unknown cause"
	case "OTHER_CAUSE":
		return "Other cause"
	case "TECHNICAL_PROBLEM":
		return "Technical problem"
	case "STRIKE":
		return "Strike or unavailable staff"
	case "DEMONSTRATION":
		return "Demonstration"
	case "ACCIDENT":
		return "Accident"
	case "HOLIDAY":
		return "Holiday"
	case "WEATHER":
		return "Weather related"
	case "MAINTENANCE":
		return "Maintenance"
	case "CONSTRUCTION":
		return "Construction work"
	case "POLICE_ACTIVITY":
		return "Police activity"
	case "MEDICAL_EMERGENCY":
		return "Medical emergency"
	case "EQUIPMENT_FAILURE":
		return "Equipment failure"
	default:
		return "Unknown cause"
	}
}

func mapGTFSRTCauseToReportType(gtfsrtCause string) string {
	switch gtfsrtCause {
	case "STRIKE", "ACCIDENT", "POLICE_ACTIVITY", "MEDICAL_EMERGENCY":
		return "incident"
	default:
		return "general"
	}
}

// Map GTFS-RT Effect → SIRI siri.Consequence Condition (minimal set)
func effectToCondition(gtfsrtEffect string) string {
	switch gtfsrtEffect {
	case "NO_SERVICE":
		return "NoService"
	case "REDUCED_SERVICE":
		return "ReducedService"
	case "SIGNIFICANT_DELAYS":
		return "SevereDelays"
	case "DETOUR":
		return "Diversion"
	default:
		return ""
	}
}
