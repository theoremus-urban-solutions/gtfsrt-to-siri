/*
Package converter transforms GTFS-Realtime data into SIRI format.

This package is data-source agnostic and config-free. It accepts pre-loaded
GTFS and GTFS-RT data along with simple options structs.

# Basic Usage

	// 1. Load GTFS and GTFS-RT (from any source)
	gtfsIndex, _ := gtfs.NewGTFSIndexFromBytes(gtfsBytes, "AGENCY")
	rt, _ := gtfsrt.NewGTFSRTWrapper(tuBytes, vpBytes, saBytes)

	// 2. Create converter options
	opts := converter.ConverterOptions{
	    AgencyID:       "AGENCY",
	    ReadIntervalMS: 30000,
	    FieldMutators: converter.FieldMutators{
	        StopPointRef: []string{"OLD_ID", "NEW_ID"},
	    },
	}

	// 3. Create converter
	conv := converter.NewConverter(gtfsIndex, rt, opts)

	// 4. Generate SIRI responses
	vm := conv.GetCompleteVehicleMonitoringResponse()
	et := conv.BuildEstimatedTimetable()
	sx := conv.BuildSituationExchange()

# SIRI Modules

Vehicle Monitoring (VM):

	vm := conv.GetCompleteVehicleMonitoringResponse()
	// Returns complete VM ServiceDelivery with all active vehicles

Estimated Timetable (ET):

	et := conv.BuildEstimatedTimetable()
	// Returns ET with recorded/estimated calls for all trips

Situation Exchange (SX):

	sx := conv.BuildSituationExchange()
	// Returns SX with all service alerts

# Field Mutators

Field mutators allow string replacement in SIRI references:

	opts := converter.ConverterOptions{
	    FieldMutators: converter.FieldMutators{
	        StopPointRef:   []string{"OLD_STOP_1", "NEW_STOP_1", "OLD_STOP_2", "NEW_STOP_2"},
	        OriginRef:      []string{"OLD_ORIGIN", "NEW_ORIGIN"},
	        DestinationRef: []string{"OLD_DEST", "NEW_DEST"},
	    },
	}

Format: [from1, to1, from2, to2, ...] - pairs of old/new values.

# Configuration

No config files required. Pass options directly:

	opts := converter.ConverterOptions{
	    AgencyID:       "AGENCY",        // Required for SIRI references
	    ReadIntervalMS: 30000,           // For ValidUntil calculation
	    FieldMutators:  FieldMutators{}, // Optional string replacements
	}

# Server Integration Pattern

Typical Kafka-based server:

	type Server struct {
	    gtfsIndex *gtfs.GTFSIndex // Cached GTFS index
	    opts      converter.ConverterOptions
	}

	func (s *Server) Init() error {
	    // Parse GTFS once at startup
	    gtfsBytes, _ := fetchFromMinIO("gtfs.zip")
	    s.gtfsIndex, _ = gtfs.NewGTFSIndexFromBytes(gtfsBytes, "AGENCY")

	    s.opts = converter.ConverterOptions{
	        AgencyID:       "AGENCY",
	        ReadIntervalMS: 30000,
	    }
	    return nil
	}

	func (s *Server) HandleKafkaMessage(pbBytes []byte) ([]byte, error) {
	    // Parse GTFS-RT from Kafka (fast - only current data)
	    rt, _ := gtfsrt.NewGTFSRTWrapper(pbBytes, nil, nil)

	    // Convert (reuses cached GTFS index - very fast)
	    conv := converter.NewConverter(s.gtfsIndex, rt, s.opts)
	    response := conv.GetCompleteVehicleMonitoringResponse()

	    // Format and return
	    rb := formatter.NewResponseBuilder()
	    return rb.BuildJSON(response), nil
	}

# Thread Safety

Converter instances are NOT thread-safe. Create a new converter per request.
The GTFS index and GTFS-RT wrapper can be safely shared across goroutines.

# Performance

- GTFS parsing: 500ms-2s (do once at startup)
- GTFS-RT parsing: 10-50ms per message
- Conversion: <1ms with cached GTFS index
- Formatting: 5-20ms for XML/JSON

# SIRI Compliance

This converter follows Entur's Nordic SIRI Profile:

- VM: Vehicle positions and journey progress
- ET: Stop-level arrival/departure predictions
- SX: Service alerts and disruptions

All outputs include proper codespace prefixes (agency_id) in references.
*/
package converter
