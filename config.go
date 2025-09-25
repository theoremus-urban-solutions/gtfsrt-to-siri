package gtfsrtsiri

import (
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port int `yaml:"port" validate:"gt=0"`
}

type GTFSConfig struct {
	IndexedSchedulePath string `yaml:"indexedScheduleData" validate:"omitempty"`
	IndexedSpatialPath  string `yaml:"indexedSpatialData" validate:"omitempty"`
	StaticURL           string `yaml:"staticURL" validate:"omitempty,url"`
	AgencyID            string `yaml:"agency_id" validate:"omitempty"`
}

type GTFSRTConfig struct {
	FeedURL             string `yaml:"feedURL" validate:"omitempty,url"`
	TripUpdatesURL      string `yaml:"tripUpdatesURL" validate:"omitempty,url"`
	VehiclePositionsURL string `yaml:"vehiclePositionsURL" validate:"omitempty,url"`
	ReadIntervalMS      int    `yaml:"readIntervalMS" validate:"gte=0"`
	TimeoutMS           int    `yaml:"timeoutMS" validate:"gte=0"`
}

type FieldMutators struct {
	OriginRef      []string `yaml:"OriginRef"`
	StopPointRef   []string `yaml:"StopPointRef"`
	DestinationRef []string `yaml:"DestinationRef"`
}

type ConverterConfig struct {
	FieldMutators                   FieldMutators `yaml:"fieldMutators"`
	UnscheduledTripIndicator        string        `yaml:"unscheduledTripIndicator"`
	CallDistanceAlongRouteNumDigits int           `yaml:"callDistanceAlongRouteNumOfDigits"`
	TripKeyStrategy                 string        `yaml:"tripKeyStrategy"` // raw|startDateTrip|agencyTrip|agencyStartDateTrip
}

type Feed struct {
	Name   string       `yaml:"name" validate:"required"`
	GTFS   GTFSConfig   `yaml:"gtfs" validate:"required"`
	GTFSRT GTFSRTConfig `yaml:"gtfsrt" validate:"required"`
}

type AppConfig struct {
	Server    ServerConfig    `yaml:"server" validate:"required"`
	GTFS      GTFSConfig      `yaml:"gtfs"`
	GTFSRT    GTFSRTConfig    `yaml:"gtfsrt"`
	Converter ConverterConfig `yaml:"converter"`
	Feeds     []Feed          `yaml:"feeds"`
}

var Config AppConfig

func LoadAppConfig() error {
	paths := []string{"config.yml", "./golang/config.yml"}
	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	v := validator.New()
	if err := v.Struct(cfg.Server); err != nil {
		return err
	}
	// feeds are optional; if present validate each
	for _, f := range cfg.Feeds {
		if err := v.Struct(f); err != nil {
			return err
		}
	}
	Config = cfg
	if Config.Server.Port == 0 {
		Config.Server.Port = 16181
	}
	return nil
}

func metricsHandler() http.Handler {
	return http.NewServeMux() // placeholder; replaced with promhttp.Handler() later
}

// SelectFeed chooses a feed by name; fallback to first; if none, use top-level GTFS/GTFSRT.
func SelectFeed(name string) (GTFSConfig, GTFSRTConfig) {
	if name != "" {
		for _, f := range Config.Feeds {
			if f.Name == name {
				return f.GTFS, f.GTFSRT
			}
		}
	}
	if len(Config.Feeds) > 0 {
		return Config.Feeds[0].GTFS, Config.Feeds[0].GTFSRT
	}
	return Config.GTFS, Config.GTFSRT
}
