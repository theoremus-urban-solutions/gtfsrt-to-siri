package converter

import (
	"log"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

func (c *Converter) BuildSituationExchange() siri.SituationExchange {
	alerts := c.gtfsrt.GetAlerts()
	elements := make([]siri.PtSituationElement, 0, len(alerts))
	now := c.gtfsrt.GetTimestampForFeedMessage()
	for _, a := range alerts {
		severity, effectPrefix := mapGTFSRTEffectToSIRISeverity(a.Effect)

		// Build situation number with codespace prefix
		codespace := c.opts.AgencyID
		if codespace == "" {
			codespace = "UNKNOWN"
		}
		situationNumber := codespace + ":SituationNumber:" + a.ID

		// Build localized summaries with both Cause and Effect
		causeEN, causeBG := mapGTFSRTCauseToSummaryCause(a.Cause)
		effectEN, effectBG := mapGTFSRTEffectToSIRISummaryEffect(a.Effect)

		summaries := []siri.LocalizedText{}
		if causeEN != "" || effectEN != "" {
			summaryEN := causeEN + ": " + effectEN
			summaries = append(summaries, siri.LocalizedText{Lang: "en", Text: summaryEN})
		}
		if causeBG != "" || effectBG != "" {
			summaryBG := causeBG + ": " + effectBG
			summaries = append(summaries, siri.LocalizedText{Lang: "bg", Text: summaryBG})
		}
		// Log if no summary available
		if len(summaries) == 0 {
			c.warnings.Add(WarningNoSummary, a.ID)
		}

		// Build localized descriptions
		descriptions := []siri.LocalizedText{}
		if len(a.DescriptionByLang) > 0 {
			for lang, desc := range a.DescriptionByLang {
				if desc != "" {
					descriptions = append(descriptions, siri.LocalizedText{Lang: lang, Text: desc})
				}
			}
		} else if a.Description != "" {
			// Fallback to single description with effect prefix if available
			description := a.Description
			if effectPrefix != "" {
				description = effectPrefix + ": " + a.Description
			}
			descriptions = append(descriptions, siri.LocalizedText{Lang: "en", Text: description})
		}
		// Log if no description available
		if len(descriptions) == 0 {
			c.warnings.Add(WarningNoDescription, a.ID)
		}

		// Build InfoLinks from URLs
		infoLinks := []siri.InfoLink{}
		if len(a.URLByLang) > 0 {
			for lang, url := range a.URLByLang {
				if url != "" {
					infoLinks = append(infoLinks, siri.InfoLink{Lang: lang, URL: url})
				}
			}
		}

		el := siri.PtSituationElement{
			ParticipantRef:  codespace,
			SituationNumber: situationNumber,
			SourceType:      "directReport",
			Severity:        severity,
			ReportType:      mapGTFSRTCauseToReportType(a.Cause),
			Summaries:       summaries,
			Descriptions:    descriptions,
			InfoLinks:       infoLinks,
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
				DatedVehicleJourneyRef: gtfsrt.TripKeyForConverter(tid, c.opts.AgencyID, c.gtfsrt.GetStartDateForTrip(tid)),
			}
			// LineRef with codespace prefix - try GTFS-RT first, then static GTFS (ALWAYS use plain tripID for static)
			rid := c.gtfsrt.GetRouteIDForTrip(tid)
			if rid == "" {
				rid = c.gtfs.GetRouteIDForTrip(tid)
				if rid == "" {
					c.warnings.Add(WarningNoRouteID, tid)
				}
			}
			if rid != "" {
				vj.LineRef = codespace + ":Line:" + rid
			}
			if dir := c.gtfsrt.GetRouteDirectionForTrip(tid); dir != "" {
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
			// Check if stop exists in static GTFS
			if stopName := c.gtfs.GetStopName(sid); stopName == "" {
				c.warnings.Add(WarningStopNotFound, sid)
			}
			el.Affects.StopPoints = append(el.Affects.StopPoints, siri.AffectedStopPoint{
				StopPointRef: applyFieldMutators(sid, c.opts.FieldMutators.StopPointRef),
			})
		}

		// Log if alert has no informed entities (system-wide alert)
		if len(a.TripIDs) == 0 && len(a.RouteIDs) == 0 && len(a.StopIDs) == 0 {
			log.Printf("[SX] INFO: alert %s has no informed entities (system-wide alert)", a.ID)
		}
		// Add siri.Consequences derived from GTFS-RT Effect
		if cond := effectToCondition(a.Effect); cond != "" {
			el.Consequences = []siri.Consequence{{Condition: cond}}
		}
		elements = append(elements, el)
	}

	// Log consolidated warnings
	codespace := c.opts.AgencyID
	if codespace == "" {
		codespace = "UNKNOWN"
	}
	c.warnings.LogAll("Alerts->SX", codespace)

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

func mapGTFSRTCauseToSummaryCause(gtfsrtCause string) (string, string) {
	switch gtfsrtCause {
	case "UNKNOWN_CAUSE":
		return "Unknown", "Неизвестно"
	case "OTHER_CAUSE":
		return "Other", "Друго"
	case "TECHNICAL_PROBLEM":
		return "Technical problem", "Технически проблем"
	case "STRIKE":
		return "Strike or unavailable staff", "Стачка или недостиг на персонал"
	case "DEMONSTRATION":
		return "Demonstration", "Демонстрация"
	case "ACCIDENT":
		return "Accident", "Авария"
	case "HOLIDAY":
		return "Holiday", "Праздник"
	case "WEATHER":
		return "Weather", "Лошо време"
	case "MAINTENANCE":
		return "Maintenance", "Поддръжка"
	case "CONSTRUCTION":
		return "Construction work", "Строителна дейност"
	case "POLICE_ACTIVITY":
		return "Police activity", "Полицейска дейност"
	case "MEDICAL_EMERGENCY":
		return "Medical emergency", "Медицинска авария"
	case "EQUIPMENT_FAILURE":
		return "Equipment failure", "Проблем с оборудване"
	default:
		return "Unknown", "Неизвестно"
	}
}
func mapGTFSRTEffectToSIRISummaryEffect(gtfsrtEffect string) (string, string) {
	switch gtfsrtEffect {
	case "NO_SERVICE":
		return "No service", "Не се изпълнява"
	case "REDUCED_SERVICE":
		return "Reduced service", "Понижено обслужване"
	case "SIGNIFICANT_DELAYS":
		return "Significant delays", "Значителни закъснения"
	case "DETOUR":
		return "Detour", "Отклонение"
	case "ADDITIONAL_SERVICE":
		return "Additional service", "Допълнително обслужване"
	case "MODIFIED_SERVICE":
		return "Modified service", "Модифицирано обслужване"
	case "OTHER_EFFECT":
		return "Other", "Друго"
	case "UNKNOWN_EFFECT":
		return "Unknown", "Неизвестно"
	case "STOP_MOVED":
		return "Stop moved", "Преместена спирка"
	case "NO_EFFECT":
		return "No impact", "Няма ефект"
	case "ACCESSIBILITY_ISSUE":
		return "Accessibility issue", "Проблем с достъпността"
	default:
		return "undefined", "Неизвестно"
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
