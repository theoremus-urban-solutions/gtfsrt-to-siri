package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// Config is the global application configuration
var Config AppConfig

// LoadAppConfig loads and validates the application configuration from config.yml
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
