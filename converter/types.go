package converter

// ConverterOptions contains all configuration needed for GTFS-RT to SIRI conversion.
// This struct is data-source agnostic and has no dependencies on config files.
type ConverterOptions struct {
	// AgencyID is the GTFS agency identifier used in SIRI reference formatting.
	// Required for proper SIRI references like {agency}:Line:{route_id}
	AgencyID string

	// ReadIntervalMS is the refresh interval in milliseconds.
	// Used to calculate ValidUntil timestamps in SIRI responses.
	ReadIntervalMS int64

	// FieldMutators defines string replacement rules for SIRI references.
	// Optional - leave empty if no mutations needed.
	FieldMutators FieldMutators
}

// FieldMutators defines string replacement rules for SIRI reference fields.
// Format: [from1, to1, from2, to2, ...] - pairs of old/new values.
//
// Example:
//
//	FieldMutators{
//	    StopPointRef: []string{"OLD_STOP_1", "NEW_STOP_1", "OLD_STOP_2", "NEW_STOP_2"},
//	}
//
// This would replace "OLD_STOP_1" with "NEW_STOP_1" and "OLD_STOP_2" with "NEW_STOP_2"
// in all StopPointRef fields.
type FieldMutators struct {
	// StopPointRef mutations for stop references
	StopPointRef []string

	// OriginRef mutations for origin references
	OriginRef []string

	// DestinationRef mutations for destination references
	DestinationRef []string
}
