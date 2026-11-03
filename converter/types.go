package converter

// ConverterOptions contains all configuration needed for GTFS-RT to SIRI conversion.
// This struct is data-source agnostic and has no dependencies on config files.
type ConverterOptions struct {
	// AgencyID is the GTFS agency identifier used in SIRI reference formatting.
	// Required for proper SIRI references like {agency}:Line:{route_id}
	AgencyID string

	// ReadIntervalMS sets the VM ValidUntilTime as the vehicle timestamp plus this interval; zero means 30 seconds.
	ReadIntervalMS int64
}

// DefaultReadIntervalMS is used when ReadIntervalMS is zero, because SIRI requires ValidUntilTime.
const DefaultReadIntervalMS = 30000
