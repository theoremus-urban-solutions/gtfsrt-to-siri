package config

// ServerConfig contains server configuration
type ServerConfig struct {
	Port int `yaml:"port" validate:"gt=0"`
}

// GTFSConfig contains GTFS static feed configuration
type GTFSConfig struct {
	StaticURL string `yaml:"staticURL" validate:"omitempty,url"`
	AgencyID  string `yaml:"agency_id" validate:"omitempty"`
}

// GTFSRTConfig contains GTFS-Realtime feed configuration
type GTFSRTConfig struct {
	FeedURL             string `yaml:"feedURL" validate:"omitempty,url"`
	TripUpdatesURL      string `yaml:"tripUpdatesURL" validate:"omitempty,url"`
	VehiclePositionsURL string `yaml:"vehiclePositionsURL" validate:"omitempty,url"`
	ServiceAlertsURL    string `yaml:"serviceAlertsURL" validate:"omitempty,url"`
	ReadIntervalMS      int    `yaml:"readIntervalMS" validate:"gte=0"`
	TimeoutMS           int    `yaml:"timeoutMS" validate:"gte=0"`
}

// FieldMutators contains field transformation rules
type FieldMutators struct {
	OriginRef      []string `yaml:"OriginRef"`
	StopPointRef   []string `yaml:"StopPointRef"`
	DestinationRef []string `yaml:"DestinationRef"`
}

// ConverterConfig contains converter-specific configuration
type ConverterConfig struct {
	FieldMutators                   FieldMutators `yaml:"fieldMutators"`
	UnscheduledTripIndicator        string        `yaml:"unscheduledTripIndicator"`
	CallDistanceAlongRouteNumDigits int           `yaml:"callDistanceAlongRouteNumOfDigits"`
	TripKeyStrategy                 string        `yaml:"tripKeyStrategy"` // raw|startDateTrip|agencyTrip|agencyStartDateTrip
}

// Feed represents a single GTFS feed configuration
type Feed struct {
	Name   string       `yaml:"name" validate:"required"`
	GTFS   GTFSConfig   `yaml:"gtfs" validate:"required"`
	GTFSRT GTFSRTConfig `yaml:"gtfsrt" validate:"required"`
}

// AppConfig is the root configuration structure
type AppConfig struct {
	Server    ServerConfig    `yaml:"server" validate:"required"`
	GTFS      GTFSConfig      `yaml:"gtfs"`
	GTFSRT    GTFSRTConfig    `yaml:"gtfsrt"`
	Converter ConverterConfig `yaml:"converter"`
	Feeds     []Feed          `yaml:"feeds"`
}
