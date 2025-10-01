package main

import (
	"flag"
	"fmt"
	lib "mta/gtfsrt-to-siri"
	"strings"
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

	lib.InitLogging()
	if err := lib.LoadAppConfig(); err != nil {
		panic(err)
	}

	gtfsCfg, rtCfg := lib.SelectFeed(*feedName)

	switch *mode {
	case "oneshot":
		gtfs, _ := lib.NewGTFSIndexFromConfig(gtfsCfg)
		tu := rtCfg.TripUpdatesURL
		vp := rtCfg.VehiclePositionsURL
		if *tripUpdates != "" {
			tu = *tripUpdates
		}
		if *vehiclePositions != "" {
			vp = *vehiclePositions
		}
		// Apply modules selection: disable URLs for modules not requested
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

		rt := lib.NewGTFSRTWrapper(tu, vp, alerts)
		_ = rt.Refresh()
		conv := lib.NewConverter(gtfs, rt, lib.Config)
		cache := lib.NewConverterCache(conv)
		var buf []byte
		var err error
		if *call == "et" {
			params := map[string]string{}
			if *monitoringRef != "" {
				params["monitoringref"] = *monitoringRef
			}
			if *lineRef != "" {
				params["lineref"] = *lineRef
			}
			if *directionRef != "" {
				params["directionref"] = *directionRef
			}
			buf, err = cache.GetEstimatedTimetableResponse(params, *format)
		} else if *call == "vm" {
			buf, err = cache.GetVehicleMonitoringResponse(map[string]string{}, *format)
		} else if *call == "sx" {
			buf, err = cache.GetSituationExchangeResponse(*format)
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(string(buf))
	default:
		panic("unknown mode")
	}
}
