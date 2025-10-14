package formatter

import (
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/siri"
	siritemp "github.com/theoremus-urban-solutions/transit-types/siri"
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
	b.WriteString(`<VehicleMonitoringDelivery version="`)
	b.WriteString(xmlEscape(vm.Version))
	b.WriteString(`">`)
	if vm.ResponseTimestamp != "" {
		b.WriteString("<ResponseTimestamp>")
		b.WriteString(xmlEscape(vm.ResponseTimestamp))
		b.WriteString("</ResponseTimestamp>")
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
		if va.MonitoredVehicleJourney != nil {
			writeMVJXML(b, *va.MonitoredVehicleJourney)
		}
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
	if mvj.DirectionRef != "" {
		b.WriteString("<DirectionRef>")
		b.WriteString(xmlEscape(mvj.DirectionRef))
		b.WriteString("</DirectionRef>")
	}
	if mvj.FramedVehicleJourneyRef != nil {
		b.WriteString("<FramedVehicleJourneyRef>")
		if mvj.FramedVehicleJourneyRef.DataFrameRef != "" {
			b.WriteString("<DataFrameRef>")
			b.WriteString(xmlEscape(mvj.FramedVehicleJourneyRef.DataFrameRef))
			b.WriteString("</DataFrameRef>")
		}
		if mvj.FramedVehicleJourneyRef.DatedVehicleJourneyRef != "" {
			b.WriteString("<DatedVehicleJourneyRef>")
			b.WriteString(xmlEscape(mvj.FramedVehicleJourneyRef.DatedVehicleJourneyRef))
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
	// Monitored field
	if mvj.Monitored != nil {
		b.WriteString("<Monitored>")
		if *mvj.Monitored {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
		b.WriteString("</Monitored>")
	}
	// DataSource (SIRI-VM spec: required)
	if mvj.DataSource != "" {
		b.WriteString("<DataSource>")
		b.WriteString(xmlEscape(mvj.DataSource))
		b.WriteString("</DataSource>")
	}
	// VehicleLocation
	if mvj.VehicleLocation != nil {
		b.WriteString("<VehicleLocation>")
		b.WriteString("<Longitude>")
		b.WriteString(strconv.FormatFloat(mvj.VehicleLocation.Longitude, 'f', 6, 64))
		b.WriteString("</Longitude>")
		b.WriteString("<Latitude>")
		b.WriteString(strconv.FormatFloat(mvj.VehicleLocation.Latitude, 'f', 6, 64))
		b.WriteString("</Latitude>")
		b.WriteString("</VehicleLocation>")
	}
	if mvj.Bearing != nil {
		b.WriteString("<Bearing>")
		b.WriteString(strconv.FormatFloat(*mvj.Bearing, 'f', 2, 64))
		b.WriteString("</Bearing>")
	}
	if mvj.Velocity != nil {
		b.WriteString("<Velocity>")
		b.WriteString(strconv.Itoa(*mvj.Velocity))
		b.WriteString("</Velocity>")
	}
	// Occupancy
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
	// InCongestion
	if mvj.InCongestion != nil {
		b.WriteString("<InCongestion>")
		if *mvj.InCongestion {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
		b.WriteString("</InCongestion>")
	}
	// VehicleStatus
	if mvj.VehicleStatus != "" {
		b.WriteString("<VehicleStatus>")
		b.WriteString(xmlEscape(mvj.VehicleStatus))
		b.WriteString("</VehicleStatus>")
	}
	// VehicleJourneyRef
	if mvj.VehicleJourneyRef != "" {
		b.WriteString("<VehicleJourneyRef>")
		b.WriteString(xmlEscape(mvj.VehicleJourneyRef))
		b.WriteString("</VehicleJourneyRef>")
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

func writeEstimatedTimetableXML(b *strings.Builder, et siritemp.EstimatedTimetableDelivery) {
	b.WriteString("<EstimatedTimetableDelivery")
	if et.Version != "" {
		b.WriteString(" version=\"")
		b.WriteString(xmlEscape(et.Version))
		b.WriteString("\"")
	}
	b.WriteString(">")
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
		// Order per SIRI-SX spec: CreationTime, ParticipantRef, SituationNumber, Version, Source, VersionedAtTime, Progress, ValidityPeriod, UndefinedReason, Severity, Priority, ReportType, Planned, Keywords, Summary, Description, Detail, Advice, Internal, Affects, Consequences, PublishingActions, InfoLinks
		if el.CreationTime != "" {
			b.WriteString("<CreationTime>")
			b.WriteString(xmlEscape(el.CreationTime))
			b.WriteString("</CreationTime>")
		}
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
		if el.Source != nil && el.Source.SourceType != "" {
			b.WriteString("<Source>")
			b.WriteString("<SourceType>")
			b.WriteString(xmlEscape(el.Source.SourceType))
			b.WriteString("</SourceType>")
			b.WriteString("</Source>")
		}
		if el.Progress != "" {
			b.WriteString("<Progress>")
			b.WriteString(xmlEscape(el.Progress))
			b.WriteString("</Progress>")
		}
		for _, vp := range el.ValidityPeriod {
			b.WriteString("<ValidityPeriod>")
			if vp.StartTime != "" {
				b.WriteString("<StartTime>")
				b.WriteString(xmlEscape(vp.StartTime))
				b.WriteString("</StartTime>")
			}
			if vp.EndTime != "" {
				b.WriteString("<EndTime>")
				b.WriteString(xmlEscape(vp.EndTime))
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
		// Multi-language summaries
		for _, summary := range el.Summary {
			b.WriteString("<Summary")
			if summary.Lang != "" {
				b.WriteString(" xml:lang=\"")
				b.WriteString(xmlEscape(summary.Lang))
				b.WriteString("\"")
			}
			b.WriteString(">")
			b.WriteString(xmlEscape(summary.Text))
			b.WriteString("</Summary>")
		}
		// Multi-language descriptions
		for _, desc := range el.Description {
			b.WriteString("<Description")
			if desc.Lang != "" {
				b.WriteString(" xml:lang=\"")
				b.WriteString(xmlEscape(desc.Lang))
				b.WriteString("\"")
			}
			b.WriteString(">")
			b.WriteString(xmlEscape(desc.Text))
			b.WriteString("</Description>")
		}
		// Affects block
		if el.Affects != nil {
			b.WriteString("<Affects>")
			// Networks > AffectedNetwork > AffectedLine
			if el.Affects.Networks != nil && len(el.Affects.Networks.AffectedNetwork) > 0 {
				b.WriteString("<Networks>")
				for _, network := range el.Affects.Networks.AffectedNetwork {
					b.WriteString("<AffectedNetwork>")
					if network.NetworkRef != "" {
						b.WriteString("<NetworkRef>")
						b.WriteString(xmlEscape(network.NetworkRef))
						b.WriteString("</NetworkRef>")
					}
					if network.AffectedLines != nil {
						for _, line := range network.AffectedLines.AffectedLine {
							b.WriteString("<AffectedLine>")
							if line.LineRef != "" {
								b.WriteString("<LineRef>")
								b.WriteString(xmlEscape(line.LineRef))
								b.WriteString("</LineRef>")
							}
							b.WriteString("</AffectedLine>")
						}
					}
					b.WriteString("</AffectedNetwork>")
				}
				b.WriteString("</Networks>")
			}
			// VehicleJourneys
			if el.Affects.VehicleJourneys != nil && len(el.Affects.VehicleJourneys.AffectedVehicleJourney) > 0 {
				for _, vj := range el.Affects.VehicleJourneys.AffectedVehicleJourney {
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
					b.WriteString("</VehicleJourney>")
				}
			}
			// StopPoints (at Affects level for stop-only alerts)
			if el.Affects.StopPoints != nil && len(el.Affects.StopPoints.AffectedStopPoint) > 0 {
				b.WriteString("<StopPoints>")
				for _, sp := range el.Affects.StopPoints.AffectedStopPoint {
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
		}
		// InfoLinks block
		if len(el.InfoLinks) > 0 {
			b.WriteString("<InfoLinks>")
			for _, link := range el.InfoLinks {
				b.WriteString("<InfoLink>")
				if link.Uri != "" {
					b.WriteString("<Uri>")
					b.WriteString(xmlEscape(link.Uri))
					b.WriteString("</Uri>")
				}
				// Label with language
				for _, label := range link.Label {
					b.WriteString("<Label")
					if label.Lang != "" {
						b.WriteString(" xml:lang=\"")
						b.WriteString(xmlEscape(label.Lang))
						b.WriteString("\"")
					}
					b.WriteString(">")
					b.WriteString(xmlEscape(label.Text))
					b.WriteString("</Label>")
				}
				b.WriteString("</InfoLink>")
			}
			b.WriteString("</InfoLinks>")
		}
		// Consequences block
		if el.Consequences != nil && len(el.Consequences.Consequence) > 0 {
			b.WriteString("<Consequences>")
			for _, c := range el.Consequences.Consequence {
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
