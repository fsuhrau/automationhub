package cmd

import (
	"fmt"
	"net"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/endpoints"
	"github.com/fsuhrau/automationhub/utils"
	"github.com/spf13/viper"
)

// loadServiceConfig reads the service config from viper, the only place the
// config sources (file, environment) are read.
func loadServiceConfig() (config.Service, error) {
	var cfg config.Service
	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = viper.GetString("log")
	}
	return cfg, nil
}

func getHostIP(cfg config.Service) net.IP {
	var hostIP net.IP
	if len(cfg.HostIP) > 0 {
		hostIP = net.ParseIP(cfg.HostIP)
	}

	if hostIP == nil {
		hostIP = utils.GetOutboundIP()
	}
	return hostIP
}

type endpointAdder interface {
	AddEndpoint(endpoints.ServiceEndpoint) error
}

// addEndpoints registers the endpoints in order and stops at the first failure.
func addEndpoints(server endpointAdder, eps ...endpoints.ServiceEndpoint) error {
	for _, ep := range eps {
		if err := server.AddEndpoint(ep); err != nil {
			return fmt.Errorf("register endpoint: %w", err)
		}
	}
	return nil
}
