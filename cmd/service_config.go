package cmd

import (
	"fmt"
	"os"

	"github.com/fsuhrau/automationhub/config"
	"github.com/spf13/viper"
)

// loadServiceConfig reads the service config from viper, the only place the
// config sources (file, environment) are read. Folders the config leaves out
// are the ones below the home of the user.
func loadServiceConfig() (config.Service, error) {
	var cfg config.Service
	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = viper.GetString("log")
	}
	cfg.Paths = cfg.Paths.WithDefaults(config.DefaultPaths(os.Getenv("HOME")))
	return cfg, nil
}
