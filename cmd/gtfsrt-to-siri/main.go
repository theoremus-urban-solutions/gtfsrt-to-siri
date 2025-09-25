package main

import (
	"flag"
	"fmt"
	lib "mta/gtfsrt-to-siri"
)

func main() {
	mode := flag.String("mode", "oneshot", "oneshot")
	format := flag.String("format", "json", "json|xml")
	call := flag.String("call", "vm", "vm|sm")
	feedName := flag.String("feed", "", "feed name from config.feeds[]")
	tripUpdates := flag.String("tripUpdates", "", "GTFS-RT TripUpdates URL (overrides config)")
	vehiclePositions := flag.String("vehiclePositions", "", "GTFS-RT VehiclePositions URL (overrides config)")
	monitoringRef := flag.String("monitoringRef", "", "StopMonitoring MonitoringRef (stop_id)")
	maxOnward := flag.Int("maxOnward", -1, "MaximumNumberOfCallsOnwards")
	lineRef := flag.String("lineRef", "", "LineRef filter (route or AGENCY_route)")
	directionRef := flag.String("directionRef", "", "DirectionRef filter (0|1)")
	maximumStopVisits := flag.Int("maximumStopVisits", -1, "MaximumStopVisits (SM selection)")
	minimumStopVisitsPerLine := flag.Int("minimumStopVisitsPerLine", -1, "MinimumStopVisitsPerLine (SM selection)")
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
		rt := lib.NewGTFSRTWrapper(tu, vp)
		_ = rt.Refresh()
		conv := lib.NewConverter(gtfs, rt, lib.Config)
		cache := lib.NewConverterCache(conv)
		var buf []byte
		var err error
		if *call == "sm" {
			ref := *monitoringRef
			if ref == "" {
				stops := gtfs.GetAllStops()
				if len(stops) > 0 {
					ref = stops[0]
				}
			}
			params := map[string]string{"monitoringref": ref}
			if *lineRef != "" {
				params["lineref"] = *lineRef
			}
			if *directionRef != "" {
				params["directionref"] = *directionRef
			}
			if *maxOnward >= 0 {
				params["maximumnumberofcallsonwards"] = fmt.Sprintf("%d", *maxOnward)
			}
			if *maximumStopVisits >= 0 {
				params["maximumstopvisits"] = fmt.Sprintf("%d", *maximumStopVisits)
			}
			if *minimumStopVisitsPerLine >= 0 {
				params["minimumstopvisitsperline"] = fmt.Sprintf("%d", *minimumStopVisitsPerLine)
			}
			buf, err = cache.GetStopMonitoringResponse(params, *format)
		} else {
			buf, err = cache.GetVehicleMonitoringResponse(map[string]string{}, *format)
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(string(buf))
	default:
		panic("unknown mode")
	}
}
