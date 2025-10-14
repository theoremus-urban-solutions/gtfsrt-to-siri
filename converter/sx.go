package converter

import (
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

		summaries := []siri.NaturalLanguageString{}
		if causeEN != "" || effectEN != "" {
			summaryEN := causeEN + ": " + effectEN
			summaries = append(summaries, siri.NaturalLanguageString{Lang: "en", Text: summaryEN})
		}
		if causeBG != "" || effectBG != "" {
			summaryBG := causeBG + ": " + effectBG
			summaries = append(summaries, siri.NaturalLanguageString{Lang: "bg", Text: summaryBG})
		}

		// Build localized descriptions
		descriptions := []siri.NaturalLanguageString{}
		if len(a.DescriptionByLang) > 0 {
			for lang, desc := range a.DescriptionByLang {
				if desc != "" {
					descriptions = append(descriptions, siri.NaturalLanguageString{Lang: lang, Text: desc})
				}
			}
		} else if a.Description != "" {
			// Fallback to single description with effect prefix if available
			description := a.Description
			if effectPrefix != "" {
				description = effectPrefix + ": " + a.Description
			}
			descriptions = append(descriptions, siri.NaturalLanguageString{Lang: "en", Text: description})
		}

		// Build InfoLinks from URLs
		infoLinks := []siri.InfoLink{}
		if len(a.URLByLang) > 0 {
			for lang, url := range a.URLByLang {
				if url != "" {
					infoLinks = append(infoLinks, siri.InfoLink{
						Uri:   url,
						Label: []siri.NaturalLanguageString{{Lang: lang, Text: lang}},
					})
				}
			}
		}

		// Build ValidityPeriod
		validityPeriods := []siri.ValidityPeriod{}
		if a.Start > 0 || a.End > 0 {
			vp := siri.ValidityPeriod{}
			if a.Start > 0 {
				vp.StartTime = utils.Iso8601FromUnixSeconds(a.Start)
			}
			if a.End > 0 {
				vp.EndTime = utils.Iso8601FromUnixSeconds(a.End)
			}
			validityPeriods = append(validityPeriods, vp)
		}

		// Build Source
		source := &siri.SituationSource{
			SourceType: "directReport",
		}

		// Determine Progress based on validity period
		progress := "open"
		if a.End > 0 && a.End < now {
			progress = "closed"
		}

		el := siri.PtSituationElement{
			CreationTime:    utils.Iso8601FromUnixSeconds(now),
			ParticipantRef:  codespace,
			SituationNumber: situationNumber,
			Source:          source,
			Progress:        progress,
			ValidityPeriod:  validityPeriods,
			Severity:        severity,
			ReportType:      mapGTFSRTCauseToReportType(a.Cause),
			Summary:         summaries,
			Description:     descriptions,
			InfoLinks:       infoLinks,
		}
		// Build Affects structure based on GTFS-RT informed_entity
		affects := &siri.Affects{}

		// Build VehicleJourneys for trip-level alerts
		var vehicleJourneys []siri.AffectedVehicleJourney
		seenTrips := map[string]bool{}
		for _, tid := range a.TripIDs {
			if seenTrips[tid] {
				continue
			}
			seenTrips[tid] = true
			vj := siri.AffectedVehicleJourney{
				DatedVehicleJourneyRef: gtfsrt.TripKeyForConverter(tid, c.opts.AgencyID, c.gtfsrt.GetStartDateForTrip(tid)),
			}
			// LineRef with codespace prefix
			if rid := c.gtfsrt.GetRouteIDForTrip(tid); rid != "" {
				vj.LineRef = codespace + ":Line:" + rid
			}
			vehicleJourneys = append(vehicleJourneys, vj)
		}
		if len(vehicleJourneys) > 0 {
			affects.VehicleJourneys = &siri.AffectedVehicleJourneys{
				AffectedVehicleJourney: vehicleJourneys,
			}
		}

		// Build Networks > AffectedLine for route-level alerts
		if len(a.RouteIDs) > 0 {
			var affectedLines []siri.AffectedLine
			for _, rid := range a.RouteIDs {
				affectedLine := siri.AffectedLine{
					LineRef: codespace + ":Line:" + rid,
				}
				affectedLines = append(affectedLines, affectedLine)
			}
			network := siri.AffectedNetwork{
				NetworkRef: codespace + ":Network:" + codespace,
				AffectedLines: &siri.AffectedLines{
					AffectedLine: affectedLines,
				},
			}
			affects.Networks = &siri.AffectedNetworks{
				AffectedNetwork: []siri.AffectedNetwork{network},
			}
		}

		// Build StopPoints for stop-only alerts
		var stopPoints []siri.AffectedStopPoint
		for _, sid := range a.StopIDs {
			stopPoints = append(stopPoints, siri.AffectedStopPoint{
				StopPointRef: applyFieldMutators(sid, c.opts.FieldMutators.StopPointRef),
			})
		}
		if len(stopPoints) > 0 {
			affects.StopPoints = &siri.AffectedStopPoints{
				AffectedStopPoint: stopPoints,
			}
		}

		if affects.Networks != nil || affects.StopPoints != nil || affects.VehicleJourneys != nil {
			el.Affects = affects
		}

		// Add Consequences derived from GTFS-RT Effect
		if cond := effectToCondition(a.Effect); cond != "" {
			el.Consequences = &siri.Consequences{
				Consequence: []siri.Consequence{{Condition: cond}},
			}
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
