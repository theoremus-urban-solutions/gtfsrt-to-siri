package gtfsrtsiri

import (
	"encoding/json"
	"strconv"
	"strings"
)

type responseBuilder struct{}

func newResponseBuilder() *responseBuilder { return &responseBuilder{} }

func (rb *responseBuilder) BuildJSON(res *SiriResponse) []byte {
	b, _ := json.Marshal(res)
	return b
}

func (rb *responseBuilder) BuildXML(res *SiriResponse) []byte {
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
	// VehicleMonitoringDelivery (support multiple deliveries)
	for _, vm := range sd.VehicleMonitoringDelivery {
		writeVehicleMonitoringXML(&b, vm)
	}
	// StopMonitoringDelivery
	for _, sm := range sd.StopMonitoringDelivery {
		writeStopMonitoringXML(&b, sm)
	}
	// SituationExchangeDelivery
	for _, sx := range sd.SituationExchangeDelivery {
		writeSituationExchangeXML(&b, sx)
	}
	b.WriteString("</ServiceDelivery>")
	b.WriteString("</Siri>")
	return []byte(b.String())
}

func writeVehicleMonitoringXML(b *strings.Builder, vm VehicleMonitoring) {
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
		writeMVJXML(b, va.MonitoredVehicleJourney)
		b.WriteString("</VehicleActivity>")
	}
	b.WriteString("</VehicleMonitoringDelivery>")
}

func writeMVJXML(b *strings.Builder, mvj MonitoredVehicleJourney) {
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
	if fr, ok := mvj.FramedVehicleJourneyRef.(FramedVehicleJourneyRef); ok {
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
	if mvj.JourneyPatternRef != "" {
		b.WriteString("<JourneyPatternRef>")
		b.WriteString(xmlEscape(mvj.JourneyPatternRef))
		b.WriteString("</JourneyPatternRef>")
	}
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
	if loc, ok := mvj.VehicleLocation.(VehicleLocation); ok {
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
	if mvj.VehicleRef != "" {
		b.WriteString("<VehicleRef>")
		b.WriteString(xmlEscape(mvj.VehicleRef))
		b.WriteString("</VehicleRef>")
	}
	writeOnwardCallsXML(b, mvj.OnwardCalls)
	b.WriteString("</MonitoredVehicleJourney>")
}

func writeOnwardCallsXML(b *strings.Builder, oc any) {
	if oc == nil {
		return
	}
	m, ok := oc.(map[string]any)
	if !ok {
		return
	}
	val, ok := m["OnwardCall"]
	if !ok {
		return
	}
	switch list := val.(type) {
	case []SiriCall:
		if len(list) == 0 {
			return
		}
		b.WriteString("<OnwardCalls>")
		for _, c := range list {
			writeCallXML(b, "OnwardCall", c)
		}
		b.WriteString("</OnwardCalls>")
	case []any:
		if len(list) == 0 {
			return
		}
		b.WriteString("<OnwardCalls>")
		for _, v := range list {
			if c, ok := v.(SiriCall); ok {
				writeCallXML(b, "OnwardCall", c)
			}
		}
		b.WriteString("</OnwardCalls>")
	}
}

func writeStopMonitoringXML(b *strings.Builder, sm StopMonitoring) {
	b.WriteString("<StopMonitoringDelivery>")
	if sm.ResponseTimestamp != "" {
		b.WriteString("<ResponseTimestamp>")
		b.WriteString(xmlEscape(sm.ResponseTimestamp))
		b.WriteString("</ResponseTimestamp>")
	}
	for _, v := range sm.MonitoredStopVisit {
		b.WriteString("<MonitoredStopVisit>")
		if v.RecordedAtTime != "" {
			b.WriteString("<RecordedAtTime>")
			b.WriteString(xmlEscape(v.RecordedAtTime))
			b.WriteString("</RecordedAtTime>")
		}
		if v.MonitoringRef != "" {
			b.WriteString("<MonitoringRef>")
			b.WriteString(xmlEscape(v.MonitoringRef))
			b.WriteString("</MonitoringRef>")
		}
		writeMVJXML(b, v.MonitoredVehicleJourney)
		writeCallXML(b, "MonitoredCall", v.MonitoredCall)
		b.WriteString("</MonitoredStopVisit>")
	}
	b.WriteString("</StopMonitoringDelivery>")
}

func writeCallXML(b *strings.Builder, tag string, c SiriCall) {
	b.WriteString("<" + tag + ">")
	if c.ExpectedArrivalTime != "" {
		b.WriteString("<ExpectedArrivalTime>")
		b.WriteString(xmlEscape(c.ExpectedArrivalTime))
		b.WriteString("</ExpectedArrivalTime>")
	}
	if c.ExpectedDepartureTime != "" {
		b.WriteString("<ExpectedDepartureTime>")
		b.WriteString(xmlEscape(c.ExpectedDepartureTime))
		b.WriteString("</ExpectedDepartureTime>")
	}
	if c.StopPointRef != "" {
		b.WriteString("<StopPointRef>")
		b.WriteString(xmlEscape(c.StopPointRef))
		b.WriteString("</StopPointRef>")
	}
	if c.StopPointName != "" {
		b.WriteString("<StopPointName>")
		b.WriteString(xmlEscape(c.StopPointName))
		b.WriteString("</StopPointName>")
	}
	b.WriteString("<VisitNumber>")
	b.WriteString(strconv.Itoa(c.VisitNumber))
	b.WriteString("</VisitNumber>")
	// Extensions
	b.WriteString("<Extensions><Distances>")
	if c.Extensions.Distances.PresentableDistance != "" {
		b.WriteString("<PresentableDistance>")
		b.WriteString(xmlEscape(c.Extensions.Distances.PresentableDistance))
		b.WriteString("</PresentableDistance>")
	}
	if c.Extensions.Distances.DistanceFromCall != nil {
		b.WriteString("<DistanceFromCall>")
		b.WriteString(strconv.FormatFloat(*c.Extensions.Distances.DistanceFromCall, 'f', -1, 64))
		b.WriteString("</DistanceFromCall>")
	}
	b.WriteString("<StopsFromCall>")
	b.WriteString(strconv.Itoa(c.Extensions.Distances.StopsFromCall))
	b.WriteString("</StopsFromCall>")
	b.WriteString("<CallDistanceAlongRoute>")
	b.WriteString(strconv.FormatFloat(c.Extensions.Distances.CallDistanceAlongRoute, 'f', -1, 64))
	b.WriteString("</CallDistanceAlongRoute>")
	b.WriteString("</Distances></Extensions>")
	b.WriteString("</" + tag + ">")
}

func writeSituationExchangeXML(b *strings.Builder, sx SituationExchange) {
	// Expect sx.Situations to be []PtSituationElement
	list, ok := sx.Situations.([]PtSituationElement)
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
		if el.SituationNumber != "" {
			b.WriteString("<SituationNumber>")
			b.WriteString(xmlEscape(el.SituationNumber))
			b.WriteString("</SituationNumber>")
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
		if el.Severity != "" {
			b.WriteString("<Severity>")
			b.WriteString(xmlEscape(el.Severity))
			b.WriteString("</Severity>")
		}
		if el.Cause != "" {
			b.WriteString("<Cause>")
			b.WriteString(xmlEscape(el.Cause))
			b.WriteString("</Cause>")
		}
		if el.Effect != "" {
			b.WriteString("<Effect>")
			b.WriteString(xmlEscape(el.Effect))
			b.WriteString("</Effect>")
		}
		if el.PublicationWindow.StartTime != "" || el.PublicationWindow.EndTime != "" {
			b.WriteString("<PublicationWindow>")
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
			b.WriteString("</PublicationWindow>")
		}
		// Affects block
		b.WriteString("<Affects>")
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
		if len(el.Affects.Routes) > 0 {
			for _, r := range el.Affects.Routes {
				b.WriteString("<Network>")
				b.WriteString("<AffectedRoute>")
				if r.LineRef != "" {
					b.WriteString("<LineRef>")
					b.WriteString(xmlEscape(r.LineRef))
					b.WriteString("</LineRef>")
				}
				b.WriteString("</AffectedRoute>")
				b.WriteString("</Network>")
			}
		}
		if len(el.Affects.StopPoints) > 0 {
			for _, sp := range el.Affects.StopPoints {
				b.WriteString("<StopPoints>")
				b.WriteString("<AffectedStopPoint>")
				if sp.StopPointRef != "" {
					b.WriteString("<StopPointRef>")
					b.WriteString(xmlEscape(sp.StopPointRef))
					b.WriteString("</StopPointRef>")
				}
				b.WriteString("</AffectedStopPoint>")
				b.WriteString("</StopPoints>")
			}
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
