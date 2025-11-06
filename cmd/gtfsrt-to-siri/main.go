package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/config"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/converter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/formatter"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfs"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/gtfsrt"
	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `GTFS-Realtime to SIRI Converter
Converts GTFS-RT feeds to SIRI format (Vehicle Monitoring, Estimated Timetable, Situation Exchange)

Usage:
  gtfsrt-to-siri [URL flags] [options]
  gtfsrt-to-siri -read-from-config [options]

Examples:
  # From URLs: Vehicle Monitoring + Estimated Timetable
  gtfsrt-to-siri \
    -tripUpdates=https://example.com/tu \
    -vehiclePositions=https://example.com/vp

  # From URLs: Estimated Timetable only (JSON)
  gtfsrt-to-siri -tripUpdates=https://example.com/tu -format=json

  # From URLs: All formats (VM + ET + SX)
  gtfsrt-to-siri \
    -tripUpdates=https://example.com/tu \
    -vehiclePositions=https://example.com/vp \
    -alerts=https://example.com/alerts

  # From config file
  gtfsrt-to-siri -read-from-config

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Output Formats Generated:
  -tripUpdates only           →  ET (Estimated Timetable)
  -alerts only                →  SX (Situation Exchange)
  -tripUpdates + -vehiclePositions  →  VM (Vehicle Monitoring) + ET
  All three URLs              →  VM + ET + SX

Note: You must provide at least one URL flag OR use -read-from-config.
      URL flags and -read-from-config are mutually exclusive.

`)
}

func main() {
	// Core flags
	format := flag.String("format", "xml", "Output format: json or xml")
	readFromConfig := flag.Bool("read-from-config", false, "Read GTFS-RT URLs from config.yml (mutually exclusive with URL flags)")

	// URL flags (mutually exclusive with -read-from-config)
	tripUpdates := flag.String("tripUpdates", "", "TripUpdates feed URL")
	vehiclePositions := flag.String("vehiclePositions", "", "VehiclePositions feed URL")
	alerts := flag.String("alerts", "", "ServiceAlerts feed URL")

	flag.Usage = printUsage
	flag.Parse()

	utils.InitLogging()
	if err := config.LoadAppConfig(); err != nil {
		panic(err)
	}

	// Validate: either read-from-config OR at least one URL flag
	hasURLFlags := *tripUpdates != "" || *vehiclePositions != "" || *alerts != ""

	if *readFromConfig && hasURLFlags {
		fmt.Fprintf(os.Stderr, "Error: -read-from-config cannot be used with URL flags (-tripUpdates, -vehiclePositions, -alerts)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if !*readFromConfig && !hasURLFlags {
		fmt.Fprintf(os.Stderr, "Error: Must provide at least one URL flag or use -read-from-config\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Determine URLs and config based on mode
	var tu, vp, alertsURL string
	var gtfsCfg config.GTFSConfig

	if *readFromConfig {
		// Read from config (use first feed or top-level config)
		var rtCfg config.GTFSRTConfig
		gtfsCfg, rtCfg = config.SelectFeed("")
		tu = rtCfg.TripUpdatesURL
		vp = rtCfg.VehiclePositionsURL
		alertsURL = rtCfg.ServiceAlertsURL
	} else {
		// Use provided URLs and top-level GTFS config
		tu = *tripUpdates
		vp = *vehiclePositions
		alertsURL = *alerts
		gtfsCfg = config.Config.GTFS
	}

	// Fetch GTFS static data
	gtfsBytes, err := gtfs.FetchGTFSData(gtfsCfg.StaticURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch GTFS: %v", err))
	}
	gtfsIndex, err := gtfs.NewGTFSIndexFromBytes(gtfsBytes, gtfsCfg.AgencyID)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse GTFS: %v", err))
	}

	// Determine which modules are available based on URLs
	includeTU := tu != ""
	includeVP := vp != ""
	includeAlerts := alertsURL != ""

	// Fetch GTFS-RT data as raw bytes
	f := newFetcher()
	tuBytes, vpBytes, alertBytes, err := f.fetchAll(tu, vp, alertsURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch GTFS-RT: %v", err))
	}

	// Create GTFS-RT wrapper from raw bytes
	rt, err := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, alertBytes)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse GTFS-RT: %v", err))
	}

	// Create converter with options from config
	opts := converter.ConverterOptions{
		AgencyID: gtfsCfg.AgencyID,
	}
	conv := converter.NewConverter(gtfsIndex, rt, opts)
	rb := formatter.NewResponseBuilder()
	codespace := config.Config.GTFS.AgencyID

	// Generate all applicable SIRI formats based on available modules
	var outputs []string

	// VM: requires both TU and VP
	if includeTU && includeVP {
		resp := conv.GetCompleteVehicleMonitoringResponse()
		var buf []byte
		if strings.ToLower(*format) == "xml" {
			buf = rb.BuildXML(resp)
		} else {
			buf = rb.BuildJSON(resp)
		}
		outputs = append(outputs, fmt.Sprintf("<!-- Vehicle Monitoring -->\n%s", string(buf)))
	}

	// ET: requires TU (with or without VP)
	if includeTU {
		et := conv.BuildEstimatedTimetable()
		resp := formatter.WrapEstimatedTimetableResponse(et, codespace)
		var buf []byte
		if strings.ToLower(*format) == "xml" {
			buf = rb.BuildXML(resp)
		} else {
			buf = rb.BuildJSON(resp)
		}
		outputs = append(outputs, fmt.Sprintf("<!-- Estimated Timetable -->\n%s", string(buf)))
	}

	// SX: requires alerts
	if includeAlerts {
		sx := conv.BuildSituationExchange()
		timestamp := rt.GetTimestampForFeedMessage()
		resp := formatter.WrapSituationExchangeResponse(sx, timestamp, codespace)
		var buf []byte
		if strings.ToLower(*format) == "xml" {
			buf = rb.BuildXML(resp)
		} else {
			buf = rb.BuildJSON(resp)
		}
		outputs = append(outputs, fmt.Sprintf("<!-- Situation Exchange -->\n%s", string(buf)))
	}

	// Output all generated formats
	if len(outputs) == 0 {
		panic("No valid module combination. Use: tu (ET), alerts (SX), or tu,vp (VM+ET)")
	}

	for _, output := range outputs {
		fmt.Println(output)
		fmt.Println() // Separator between formats
	}
}
