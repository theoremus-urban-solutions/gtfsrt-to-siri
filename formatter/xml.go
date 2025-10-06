package formatter

import (
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
)

// BuildXML serializes a SIRI response to XML
func (rb *responseBuilder) BuildXML(res *siri.SiriResponse) []byte {
	var b strings.Builder
	b.WriteString("<Siri xmlns=\"http://www.siri.org.uk/siri\">")
	// ServiceDelivery
	sd := res.Siri.ServiceDelivery
	b.WriteString("<ServiceDelivery>")
	if sd.ResponseTimestamp != "" {
		b.WriteString("<ResponseTimestamp>")
		b.WriteString(xmlEscape(sd.ResponseTimestamp))
		b.WriteString("</ResponseTimestamp>")
	}
	if sd.ProducerRef != "" {
		b.WriteString("<ProducerRef>")
		b.WriteString(xmlEscape(sd.ProducerRef))
		b.WriteString("</ProducerRef>")
	}
	// VehicleMonitoringDelivery (support multiple deliveries)
	for _, vm := range sd.VehicleMonitoringDelivery {
		writeVehicleMonitoringXML(&b, vm)
	}
	// EstimatedTimetableDelivery
	for _, et := range sd.EstimatedTimetableDelivery {
		writeEstimatedTimetableXML(&b, et)
	}
	// SituationExchangeDelivery
	for _, sx := range sd.SituationExchangeDelivery {
		writeSituationExchangeXML(&b, sx)
	}
	b.WriteString("</ServiceDelivery>")
	b.WriteString("</Siri>")
	return []byte(b.String())
}

func writeVehicleMonitoringXML(b *strings.Builder, vm siri.VehicleMonitoring) {
	b.WriteString("<VehicleMonitoringDelivery>")
	if vm.ResponseTimestamp != "" {
		b.WriteString("<ResponseTimestamp>")
		b.WriteString(xmlEscape(vm.ResponseTimestamp))
		b.WriteString("</ResponseTimestamp>")
	}
	if vm.ValidUntil != "" {
		b.WriteString("<ValidUntil>")
		b.WriteString(xmlEscape(vm.ValidUntil))
		b.WriteString("</ValidUntil>")
	}
	for _, va := range vm.VehicleActivity {
		b.WriteString("<VehicleActivity>")
		if va.RecordedAtTime != "" {
			b.WriteString("<RecordedAtTime>")
			b.WriteString(xmlEscape(va.RecordedAtTime))
			b.WriteString("</RecordedAtTime>")
		}
		if va.ValidUntilTime != "" {
			b.WriteString("<ValidUntilTime>")
			b.WriteString(xmlEscape(va.ValidUntilTime))
			b.WriteString("</ValidUntilTime>")
		}
		writeMVJXML(b, va.MonitoredVehicleJourney)
		b.WriteString("</VehicleActivity>")
	}
	b.WriteString("</VehicleMonitoringDelivery>")
}

func writeMVJXML(b *strings.Builder, mvj siri.MonitoredVehicleJourney) {
	b.WriteString("<MonitoredVehicleJourney>")
	if mvj.LineRef != "" {
		b.WriteString("<LineRef>")
		b.WriteString(xmlEscape(mvj.LineRef))
		b.WriteString("</LineRef>")
	}
	switch v := mvj.DirectionRef.(type) {
	case string:
		if v != "" {
			b.WriteString("<DirectionRef>")
			b.WriteString(xmlEscape(v))
			b.WriteString("</DirectionRef>")
		}
	case float64:
		b.WriteString("<DirectionRef>")
		b.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		b.WriteString("</DirectionRef>")
	}
	if fr, ok := mvj.FramedVehicleJourneyRef.(siri.FramedVehicleJourneyRef); ok {
		b.WriteString("<FramedVehicleJourneyRef>")
		if fr.DataFrameRef != "" {
			b.WriteString("<DataFrameRef>")
			b.WriteString(xmlEscape(fr.DataFrameRef))
			b.WriteString("</DataFrameRef>")
		}
		if fr.DatedVehicleJourneyRef != "" {
			b.WriteString("<DatedVehicleJourneyRef>")
			b.WriteString(xmlEscape(fr.DatedVehicleJourneyRef))
			b.WriteString("</DatedVehicleJourneyRef>")
		}
		b.WriteString("</FramedVehicleJourneyRef>")
	}
	// VehicleMode (right after FramedVehicleJourneyRef)
	if mvj.VehicleMode != "" {
		b.WriteString("<VehicleMode>")
		b.WriteString(xmlEscape(mvj.VehicleMode))
		b.WriteString("</VehicleMode>")
	}
	// JourneyPatternRef - REMOVED for VM spec compliance (not in Entur spec)
	if mvj.PublishedLineName != "" {
		b.WriteString("<PublishedLineName>")
		b.WriteString(xmlEscape(mvj.PublishedLineName))
		b.WriteString("</PublishedLineName>")
	}
	if mvj.OperatorRef != "" {
		b.WriteString("<OperatorRef>")
		b.WriteString(xmlEscape(mvj.OperatorRef))
		b.WriteString("</OperatorRef>")
	}
	if mvj.OriginRef != "" {
		b.WriteString("<OriginRef>")
		b.WriteString(xmlEscape(mvj.OriginRef))
		b.WriteString("</OriginRef>")
	}
	if mvj.OriginName != "" {
		b.WriteString("<OriginName>")
		b.WriteString(xmlEscape(mvj.OriginName))
		b.WriteString("</OriginName>")
	}
	if mvj.DestinationRef != "" {
		b.WriteString("<DestinationRef>")
		b.WriteString(xmlEscape(mvj.DestinationRef))
		b.WriteString("</DestinationRef>")
	}
	if mvj.DestinationName != "" {
		b.WriteString("<DestinationName>")
		b.WriteString(xmlEscape(mvj.DestinationName))
		b.WriteString("</DestinationName>")
	}
	if mvj.OriginAimedDepartureTime != "" {
		b.WriteString("<OriginAimedDepartureTime>")
		b.WriteString(xmlEscape(mvj.OriginAimedDepartureTime))
		b.WriteString("</OriginAimedDepartureTime>")
	}
	b.WriteString("<Monitored>")
	if mvj.Monitored {
		b.WriteString("true")
	} else {
		b.WriteString("false")
	}
	b.WriteString("</Monitored>")
	// InCongestion (placed right above DataSource)
	if mvj.InCongestion != nil {
		b.WriteString("<InCongestion>")
		if *mvj.InCongestion {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
		b.WriteString("</InCongestion>")
	}
	// DataSource (SIRI-VM spec: required)
	if mvj.DataSource != "" {
		b.WriteString("<DataSource>")
		b.WriteString(xmlEscape(mvj.DataSource))
		b.WriteString("</DataSource>")
	}
	if loc, ok := mvj.VehicleLocation.(siri.VehicleLocation); ok {
		if loc.Latitude != nil || loc.Longitude != nil {
			b.WriteString("<VehicleLocation>")
			if loc.Latitude != nil {
				b.WriteString("<Latitude>")
				b.WriteString(strconv.FormatFloat(*loc.Latitude, 'f', 6, 64))
				b.WriteString("</Latitude>")
			}
			if loc.Longitude != nil {
				b.WriteString("<Longitude>")
				b.WriteString(strconv.FormatFloat(*loc.Longitude, 'f', 6, 64))
				b.WriteString("</Longitude>")
			}
			b.WriteString("</VehicleLocation>")
		}
	}
	if mvj.Bearing != nil {
		b.WriteString("<Bearing>")
		b.WriteString(strconv.FormatFloat(*mvj.Bearing, 'f', 2, 64))
		b.WriteString("</Bearing>")
	}
	// Occupancy (placed right above Delay)
	if mvj.Occupancy != "" {
		b.WriteString("<Occupancy>")
		b.WriteString(xmlEscape(mvj.Occupancy))
		b.WriteString("</Occupancy>")
	}
	// Delay (SIRI-VM spec: required)
	if mvj.Delay != "" {
		b.WriteString("<Delay>")
		b.WriteString(xmlEscape(mvj.Delay))
		b.WriteString("</Delay>")
	}
	if mvj.VehicleRef != "" {
		b.WriteString("<VehicleRef>")
		b.WriteString(xmlEscape(mvj.VehicleRef))
		b.WriteString("</VehicleRef>")
	}
	// MonitoredCall (SIRI-VM spec: current/previous stop only)
	if mvj.MonitoredCall != nil {
		b.WriteString("<MonitoredCall>")
		b.WriteString("<StopPointRef>")
		b.WriteString(xmlEscape(mvj.MonitoredCall.StopPointRef))
		b.WriteString("</StopPointRef>")
		if mvj.MonitoredCall.Order != nil {
			b.WriteString("<Order>")
			b.WriteString(strconv.Itoa(*mvj.MonitoredCall.Order))
			b.WriteString("</Order>")
		}
		if mvj.MonitoredCall.StopPointName != "" {
			b.WriteString("<StopPointName>")
			b.WriteString(xmlEscape(mvj.MonitoredCall.StopPointName))
			b.WriteString("</StopPointName>")
		}
		if mvj.MonitoredCall.VehicleAtStop != nil {
			b.WriteString("<VehicleAtStop>")
			if *mvj.MonitoredCall.VehicleAtStop {
				b.WriteString("true")
			} else {
				b.WriteString("false")
			}
			b.WriteString("</VehicleAtStop>")
		}
		b.WriteString("</MonitoredCall>")
	}
	// IsCompleteStopSequence (SIRI-VM spec: required, always false)
	b.WriteString("<IsCompleteStopSequence>")
	if mvj.IsCompleteStopSequence {
		b.WriteString("true")
	} else {
		b.WriteString("false")
	}
	b.WriteString("</IsCompleteStopSequence>")
	// OnwardCalls - REMOVED for VM spec compliance (belongs in ET)
	// writeOnwardCallsXML(b, mvj.OnwardCalls)
	b.WriteString("</MonitoredVehicleJourney>")
}

func writeEstimatedTimetableXML(b *strings.Builder, et siri.EstimatedTimetable) {
	b.WriteString("<EstimatedTimetableDelivery>")
	if et.ResponseTimestamp != "" {
		b.WriteString("<ResponseTimestamp>")
		b.WriteString(xmlEscape(et.ResponseTimestamp))
		b.WriteString("</ResponseTimestamp>")
	}
	for _, frame := range et.EstimatedJourneyVersionFrame {
		b.WriteString("<EstimatedJourneyVersionFrame>")
		if frame.RecordedAtTime != "" {
			b.WriteString("<RecordedAtTime>")
			b.WriteString(xmlEscape(frame.RecordedAtTime))
			b.WriteString("</RecordedAtTime>")
		}
		for _, journey := range frame.EstimatedVehicleJourney {
			b.WriteString("<EstimatedVehicleJourney>")
			if journey.RecordedAtTime != "" {
				b.WriteString("<RecordedAtTime>")
				b.WriteString(xmlEscape(journey.RecordedAtTime))
				b.WriteString("</RecordedAtTime>")
			}
			if journey.LineRef != "" {
				b.WriteString("<LineRef>")
				b.WriteString(xmlEscape(journey.LineRef))
				b.WriteString("</LineRef>")
			}
			if journey.DirectionRef != "" {
				b.WriteString("<DirectionRef>")
				b.WriteString(xmlEscape(journey.DirectionRef))
				b.WriteString("</DirectionRef>")
			}
			// siri.FramedVehicleJourneyRef
			if journey.FramedVehicleJourneyRef.DatedVehicleJourneyRef != "" {
				b.WriteString("<FramedVehicleJourneyRef>")
				if journey.FramedVehicleJourneyRef.DataFrameRef != "" {
					b.WriteString("<DataFrameRef>")
					b.WriteString(xmlEscape(journey.FramedVehicleJourneyRef.DataFrameRef))
					b.WriteString("</DataFrameRef>")
				}
				b.WriteString("<DatedVehicleJourneyRef>")
				b.WriteString(xmlEscape(journey.FramedVehicleJourneyRef.DatedVehicleJourneyRef))
				b.WriteString("</DatedVehicleJourneyRef>")
				b.WriteString("</FramedVehicleJourneyRef>")
			}
			if journey.VehicleMode != "" {
				b.WriteString("<VehicleMode>")
				b.WriteString(xmlEscape(journey.VehicleMode))
				b.WriteString("</VehicleMode>")
			}
			if journey.OriginName != "" {
				b.WriteString("<OriginName>")
				b.WriteString(xmlEscape(journey.OriginName))
				b.WriteString("</OriginName>")
			}
			if journey.DestinationName != "" {
				b.WriteString("<DestinationName>")
				b.WriteString(xmlEscape(journey.DestinationName))
				b.WriteString("</DestinationName>")
			}
			if journey.OperatorRef != "" {
				b.WriteString("<OperatorRef>")
				b.WriteString(xmlEscape(journey.OperatorRef))
				b.WriteString("</OperatorRef>")
			}
			b.WriteString("<Monitored>")
			if journey.Monitored {
				b.WriteString("true")
			} else {
				b.WriteString("false")
			}
			b.WriteString("</Monitored>")
			if journey.DataSource != "" {
				b.WriteString("<DataSource>")
				b.WriteString(xmlEscape(journey.DataSource))
				b.WriteString("</DataSource>")
			}
			// VehicleRef - REMOVED from ET (only in VM per spec)
			// RecordedCalls
			if len(journey.RecordedCalls) > 0 {
				b.WriteString("<RecordedCalls>")
				for _, call := range journey.RecordedCalls {
					b.WriteString("<RecordedCall>")
					if call.StopPointRef != "" {
						b.WriteString("<StopPointRef>")
						b.WriteString(xmlEscape(call.StopPointRef))
						b.WriteString("</StopPointRef>")
					}
					if call.Order > 0 {
						b.WriteString("<Order>")
						b.WriteString(strconv.Itoa(call.Order))
						b.WriteString("</Order>")
					}
					if call.StopPointName != "" {
						b.WriteString("<StopPointName>")
						b.WriteString(xmlEscape(call.StopPointName))
						b.WriteString("</StopPointName>")
					}
					// Always write Cancellation and RequestStop
					b.WriteString("<Cancellation>")
					if call.Cancellation {
						b.WriteString("true")
					} else {
						b.WriteString("false")
					}
					b.WriteString("</Cancellation>")
					b.WriteString("<RequestStop>")
					if call.RequestStop {
						b.WriteString("true")
					} else {
						b.WriteString("false")
					}
					b.WriteString("</RequestStop>")
					if call.AimedArrivalTime != "" {
						b.WriteString("<AimedArrivalTime>")
						b.WriteString(xmlEscape(call.AimedArrivalTime))
						b.WriteString("</AimedArrivalTime>")
					}
					if call.ActualArrivalTime != "" {
						b.WriteString("<ActualArrivalTime>")
						b.WriteString(xmlEscape(call.ActualArrivalTime))
						b.WriteString("</ActualArrivalTime>")
					}
					if call.AimedDepartureTime != "" {
						b.WriteString("<AimedDepartureTime>")
						b.WriteString(xmlEscape(call.AimedDepartureTime))
						b.WriteString("</AimedDepartureTime>")
					}
					if call.ActualDepartureTime != "" {
						b.WriteString("<ActualDepartureTime>")
						b.WriteString(xmlEscape(call.ActualDepartureTime))
						b.WriteString("</ActualDepartureTime>")
					}
					b.WriteString("</RecordedCall>")
				}
				b.WriteString("</RecordedCalls>")
			}
			// EstimatedCalls
			if len(journey.EstimatedCalls) > 0 {
				b.WriteString("<EstimatedCalls>")
				for _, call := range journey.EstimatedCalls {
					b.WriteString("<EstimatedCall>")
					if call.StopPointRef != "" {
						b.WriteString("<StopPointRef>")
						b.WriteString(xmlEscape(call.StopPointRef))
						b.WriteString("</StopPointRef>")
					}
					if call.Order > 0 {
						b.WriteString("<Order>")
						b.WriteString(strconv.Itoa(call.Order))
						b.WriteString("</Order>")
					}
					if call.StopPointName != "" {
						b.WriteString("<StopPointName>")
						b.WriteString(xmlEscape(call.StopPointName))
						b.WriteString("</StopPointName>")
					}
					// Always write Cancellation and RequestStop
					b.WriteString("<Cancellation>")
					if call.Cancellation {
						b.WriteString("true")
					} else {
						b.WriteString("false")
					}
					b.WriteString("</Cancellation>")
					b.WriteString("<RequestStop>")
					if call.RequestStop {
						b.WriteString("true")
					} else {
						b.WriteString("false")
					}
					b.WriteString("</RequestStop>")
					if call.AimedArrivalTime != "" {
						b.WriteString("<AimedArrivalTime>")
						b.WriteString(xmlEscape(call.AimedArrivalTime))
						b.WriteString("</AimedArrivalTime>")
					}
					if call.ExpectedArrivalTime != "" {
						b.WriteString("<ExpectedArrivalTime>")
						b.WriteString(xmlEscape(call.ExpectedArrivalTime))
						b.WriteString("</ExpectedArrivalTime>")
					}
					if call.ArrivalStatus != "" {
						b.WriteString("<ArrivalStatus>")
						b.WriteString(xmlEscape(call.ArrivalStatus))
						b.WriteString("</ArrivalStatus>")
					}
					if call.AimedDepartureTime != "" {
						b.WriteString("<AimedDepartureTime>")
						b.WriteString(xmlEscape(call.AimedDepartureTime))
						b.WriteString("</AimedDepartureTime>")
					}
					if call.ExpectedDepartureTime != "" {
						b.WriteString("<ExpectedDepartureTime>")
						b.WriteString(xmlEscape(call.ExpectedDepartureTime))
						b.WriteString("</ExpectedDepartureTime>")
					}
					if call.DepartureStatus != "" {
						b.WriteString("<DepartureStatus>")
						b.WriteString(xmlEscape(call.DepartureStatus))
						b.WriteString("</DepartureStatus>")
					}
					b.WriteString("</EstimatedCall>")
				}
				b.WriteString("</EstimatedCalls>")
			}
			b.WriteString("<IsCompleteStopSequence>")
			if journey.IsCompleteStopSequence {
				b.WriteString("true")
			} else {
				b.WriteString("false")
			}
			b.WriteString("</IsCompleteStopSequence>")
			b.WriteString("</EstimatedVehicleJourney>")
		}
		b.WriteString("</EstimatedJourneyVersionFrame>")
	}
	b.WriteString("</EstimatedTimetableDelivery>")
}

func writeSituationExchangeXML(b *strings.Builder, sx siri.SituationExchange) {
	list, ok := sx.Situations.([]siri.PtSituationElement)
	if !ok {
		// nothing to write
		return
	}
	if len(list) == 0 {
		return
	}
	b.WriteString("<SituationExchangeDelivery>")
	b.WriteString("<Situations>")
	for _, el := range list {
		b.WriteString("<PtSituationElement>")
		// Order: ParticipantRef, SituationNumber, Source, Progress, ValidityPeriod (PublicationWindow), UndefinedReason, Severity, ReportType, Summary, Description, Affects
		if el.ParticipantRef != "" {
			b.WriteString("<ParticipantRef>")
			b.WriteString(xmlEscape(el.ParticipantRef))
			b.WriteString("</ParticipantRef>")
		}
		if el.SituationNumber != "" {
			b.WriteString("<SituationNumber>")
			b.WriteString(xmlEscape(el.SituationNumber))
			b.WriteString("</SituationNumber>")
		}
		if el.SourceType != "" {
			b.WriteString("<Source>")
			b.WriteString("<SourceType>")
			b.WriteString(xmlEscape(el.SourceType))
			b.WriteString("</SourceType>")
			b.WriteString("</Source>")
		}
		if el.Progress != "" {
			b.WriteString("<Progress>")
			b.WriteString(xmlEscape(el.Progress))
			b.WriteString("</Progress>")
		}
		if el.PublicationWindow.StartTime != "" || el.PublicationWindow.EndTime != "" {
			b.WriteString("<ValidityPeriod>")
			if el.PublicationWindow.StartTime != "" {
				b.WriteString("<StartTime>")
				b.WriteString(xmlEscape(el.PublicationWindow.StartTime))
				b.WriteString("</StartTime>")
			}
			if el.PublicationWindow.EndTime != "" {
				b.WriteString("<EndTime>")
				b.WriteString(xmlEscape(el.PublicationWindow.EndTime))
				b.WriteString("</EndTime>")
			}
			b.WriteString("</ValidityPeriod>")
		}
		b.WriteString("<UndefinedReason/>")
		if el.Severity != "" {
			b.WriteString("<Severity>")
			b.WriteString(xmlEscape(el.Severity))
			b.WriteString("</Severity>")
		}
		if el.ReportType != "" {
			b.WriteString("<ReportType>")
			b.WriteString(xmlEscape(el.ReportType))
			b.WriteString("</ReportType>")
		}
		if el.Summary != "" {
			b.WriteString("<Summary>")
			b.WriteString(xmlEscape(el.Summary))
			b.WriteString("</Summary>")
		}
		if el.Description != "" {
			b.WriteString("<Description>")
			b.WriteString(xmlEscape(el.Description))
			b.WriteString("</Description>")
		}
		// Affects block
		b.WriteString("<Affects>")
		// Networks > siri.AffectedNetwork > siri.AffectedLine > siri.AffectedRoute
		if len(el.Affects.Networks) > 0 {
			b.WriteString("<Networks>")
			for _, network := range el.Affects.Networks {
				b.WriteString("<AffectedNetwork>")
				if network.NetworkRef != "" {
					b.WriteString("<NetworkRef>")
					b.WriteString(xmlEscape(network.NetworkRef))
					b.WriteString("</NetworkRef>")
				}
				for _, line := range network.AffectedLines {
					b.WriteString("<AffectedLine>")
					if line.LineRef != "" {
						b.WriteString("<LineRef>")
						b.WriteString(xmlEscape(line.LineRef))
						b.WriteString("</LineRef>")
					}
					// AffectedRoutes with Direction and StopPoints
					if len(line.AffectedRoutes) > 0 {
						for _, route := range line.AffectedRoutes {
							b.WriteString("<AffectedRoute>")
							if route.DirectionRef != "" {
								b.WriteString("<Direction>")
								b.WriteString("<DirectionRef>")
								b.WriteString(xmlEscape(route.DirectionRef))
								b.WriteString("</DirectionRef>")
								b.WriteString("</Direction>")
							}
							if len(route.StopPoints) > 0 {
								b.WriteString("<StopPoints>")
								for _, sp := range route.StopPoints {
									b.WriteString("<AffectedStopPoint>")
									if sp.StopPointRef != "" {
										b.WriteString("<StopPointRef>")
										b.WriteString(xmlEscape(sp.StopPointRef))
										b.WriteString("</StopPointRef>")
									}
									b.WriteString("</AffectedStopPoint>")
								}
								b.WriteString("</StopPoints>")
							}
							b.WriteString("</AffectedRoute>")
						}
					}
					b.WriteString("</AffectedLine>")
				}
				b.WriteString("</AffectedNetwork>")
			}
			b.WriteString("</Networks>")
		}
		// VehicleJourneys
		if len(el.Affects.VehicleJourneys) > 0 {
			for _, vj := range el.Affects.VehicleJourneys {
				b.WriteString("<VehicleJourney>")
				if vj.DatedVehicleJourneyRef != "" {
					b.WriteString("<DatedVehicleJourneyRef>")
					b.WriteString(xmlEscape(vj.DatedVehicleJourneyRef))
					b.WriteString("</DatedVehicleJourneyRef>")
				}
				if vj.LineRef != "" {
					b.WriteString("<LineRef>")
					b.WriteString(xmlEscape(vj.LineRef))
					b.WriteString("</LineRef>")
				}
				if vj.DirectionRef != "" {
					b.WriteString("<DirectionRef>")
					b.WriteString(xmlEscape(vj.DirectionRef))
					b.WriteString("</DirectionRef>")
				}
				b.WriteString("</VehicleJourney>")
			}
		}
		// StopPoints (at Affects level for stop-only alerts)
		if len(el.Affects.StopPoints) > 0 {
			b.WriteString("<StopPoints>")
			for _, sp := range el.Affects.StopPoints {
				b.WriteString("<AffectedStopPoint>")
				if sp.StopPointRef != "" {
					b.WriteString("<StopPointRef>")
					b.WriteString(xmlEscape(sp.StopPointRef))
					b.WriteString("</StopPointRef>")
				}
				b.WriteString("</AffectedStopPoint>")
			}
			b.WriteString("</StopPoints>")
		}
		b.WriteString("</Affects>")
		// Consequences block
		if len(el.Consequences) > 0 {
			b.WriteString("<Consequences>")
			for _, c := range el.Consequences {
				b.WriteString("<Consequence>")
				if c.Condition != "" {
					b.WriteString("<Condition>")
					b.WriteString(xmlEscape(c.Condition))
					b.WriteString("</Condition>")
				}
				b.WriteString("</Consequence>")
			}
			b.WriteString("</Consequences>")
		}
		b.WriteString("</PtSituationElement>")
	}
	b.WriteString("</Situations>")
	b.WriteString("</SituationExchangeDelivery>")
}

func xmlEscape(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(s)
}
