package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/config"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/formatter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfs"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

func main() {
	mode := flag.String("mode", "oneshot", "oneshot")
	format := flag.String("format", "json", "json|xml")
	call := flag.String("call", "vm", "vm|et|sx")
	feedName := flag.String("feed", "", "feed name from config.feeds[]")
	tripUpdates := flag.String("tripUpdates", "", "GTFS-RT TripUpdates URL (overrides config)")
	vehiclePositions := flag.String("vehiclePositions", "", "GTFS-RT VehiclePositions URL (overrides config)")
	serviceAlerts := flag.String("serviceAlerts", "", "GTFS-RT ServiceAlerts URL (overrides config)")
	monitoringRef := flag.String("monitoringRef", "", "MonitoringRef (stop_id) for filtering")
	lineRef := flag.String("lineRef", "", "LineRef filter (route or AGENCY_route)")
	directionRef := flag.String("directionRef", "", "DirectionRef filter (0|1)")
	modules := flag.String("modules", "tu,vp", "Comma-separated GTFS-RT modules to fetch: tu,vp,alerts")
	flag.Parse()

	utils.InitLogging()
	if err := config.LoadAppConfig(); err != nil {
		panic(err)
	}

	gtfsCfg, rtCfg := config.SelectFeed(*feedName)

	switch *mode {
	case "oneshot":
		// Performance metrics
		totalStart := time.Now()

		// Fetch GTFS static data
		gtfsStart := time.Now()
		gtfsBytes, err := gtfs.FetchGTFSData(gtfsCfg.StaticURL)
		if err != nil {
			panic(fmt.Sprintf("Failed to fetch GTFS: %v", err))
		}
		gtfsFetchDuration := time.Since(gtfsStart)

		gtfsParseStart := time.Now()
		gtfsIndex, err := gtfs.NewGTFSIndexFromBytes(gtfsBytes, gtfsCfg.AgencyID)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse GTFS: %v", err))
		}
		gtfsParseDuration := time.Since(gtfsParseStart)

		// Determine which modules to fetch
		tu := rtCfg.TripUpdatesURL
		vp := rtCfg.VehiclePositionsURL
		if *tripUpdates != "" {
			tu = *tripUpdates
		}
		if *vehiclePositions != "" {
			vp = *vehiclePositions
		}

		includeTU, includeVP, includeAlerts := false, false, false
		{
			mset := map[string]bool{}
			for _, m := range strings.Split(*modules, ",") {
				m = strings.TrimSpace(strings.ToLower(m))
				if m != "" {
					mset[m] = true
				}
			}
			includeTU = mset["tu"]
			includeVP = mset["vp"]
			includeAlerts = mset["alerts"]
		}
		if !includeTU {
			tu = ""
		}
		if !includeVP {
			vp = ""
		}

		alerts := rtCfg.ServiceAlertsURL
		if *serviceAlerts != "" {
			alerts = *serviceAlerts
		}
		if !includeAlerts {
			alerts = ""
		}
		if *call == "sx" && !includeAlerts {
			panic("alerts module required for sx call; include via -modules=alerts")
		}

		// Fetch GTFS-RT data as raw bytes
		gtfsrtFetchStart := time.Now()
		f := newFetcher()
		tuBytes, vpBytes, alertBytes, err := f.fetchAll(tu, vp, alerts)
		if err != nil {
			panic(fmt.Sprintf("Failed to fetch GTFS-RT: %v", err))
		}
		gtfsrtFetchDuration := time.Since(gtfsrtFetchStart)

		// Create GTFS-RT wrapper from raw bytes
		gtfsrtParseStart := time.Now()
		rt, err := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, alertBytes)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse GTFS-RT: %v", err))
		}
		gtfsrtParseDuration := time.Since(gtfsrtParseStart)

		// Create converter with options from config
		opts := converter.ConverterOptions{
			AgencyID:       gtfsCfg.AgencyID,
			ReadIntervalMS: int64(rtCfg.ReadIntervalMS),
			FieldMutators: converter.FieldMutators{
				StopPointRef:   config.Config.Converter.FieldMutators.StopPointRef,
				OriginRef:      config.Config.Converter.FieldMutators.OriginRef,
				DestinationRef: config.Config.Converter.FieldMutators.DestinationRef,
			},
		}
		conv := converter.NewConverter(gtfsIndex, rt, opts)
		rb := formatter.NewResponseBuilder()

		var buf []byte
		codespace := config.Config.GTFS.AgencyID

		var conversionDuration, formattingDuration time.Duration

		switch *call {
		case "et":
			conversionStart := time.Now()
			et := conv.BuildEstimatedTimetable()
			// Apply filters if provided
			if *monitoringRef != "" || *lineRef != "" || *directionRef != "" {
				et = formatter.FilterEstimatedTimetable(et, *monitoringRef, *lineRef, *directionRef)
			}
			// Wrap in SIRI response
			resp := formatter.WrapEstimatedTimetableResponse(et, codespace)
			conversionDuration = time.Since(conversionStart)

			formattingStart := time.Now()
			if strings.ToLower(*format) == "xml" {
				buf = rb.BuildXML(resp)
			} else {
				buf = rb.BuildJSON(resp)
			}
			formattingDuration = time.Since(formattingStart)
		case "vm":
			conversionStart := time.Now()
			resp := conv.GetCompleteVehicleMonitoringResponse()
			conversionDuration = time.Since(conversionStart)

			formattingStart := time.Now()
			if strings.ToLower(*format) == "xml" {
				buf = rb.BuildXML(resp)
			} else {
				buf = rb.BuildJSON(resp)
			}
			formattingDuration = time.Since(formattingStart)
		case "sx":
			conversionStart := time.Now()
			sx := conv.BuildSituationExchange()
			timestamp := rt.GetTimestampForFeedMessage()
			resp := formatter.WrapSituationExchangeResponse(sx, timestamp, codespace)
			conversionDuration = time.Since(conversionStart)

			formattingStart := time.Now()
			if strings.ToLower(*format) == "xml" {
				buf = rb.BuildXML(resp)
			} else {
				buf = rb.BuildJSON(resp)
			}
			formattingDuration = time.Since(formattingStart)
		}

		totalDuration := time.Since(totalStart)

		// Log performance metrics
		log.Printf("Performance metrics for %s conversion:", *call)
		log.Printf("  GTFS static fetch:  %v (%.0f ms)", gtfsFetchDuration, gtfsFetchDuration.Seconds()*1000)
		log.Printf("  GTFS static parse:  %v (%.0f ms)", gtfsParseDuration, gtfsParseDuration.Seconds()*1000)
		log.Printf("  GTFS-RT fetch:      %v (%.0f ms)", gtfsrtFetchDuration, gtfsrtFetchDuration.Seconds()*1000)
		log.Printf("  GTFS-RT parse:      %v (%.0f ms)", gtfsrtParseDuration, gtfsrtParseDuration.Seconds()*1000)
		log.Printf("  Conversion:         %v (%.0f ms)", conversionDuration, conversionDuration.Seconds()*1000)
		log.Printf("  Formatting (%s):    %v (%.0f ms)", *format, formattingDuration, formattingDuration.Seconds()*1000)
		log.Printf("  Total:              %v (%.0f ms)", totalDuration, totalDuration.Seconds()*1000)
		log.Printf("  Output size:        %d bytes (%.2f KB)", len(buf), float64(len(buf))/1024.0)

		fmt.Println(string(buf))
	default:
		panic("unknown mode")
	}
}
