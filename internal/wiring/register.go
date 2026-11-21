package wiring

import (
	"fmt"
	"sort"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/endpoints"
)

// handlerAdder is the device manager as the wiring uses it.
type handlerAdder interface {
	AddHandler(device.Handler)
}

func addHandlers(to handlerAdder, handlers []device.Handler) {
	for _, h := range handlers {
		to.AddHandler(h)
	}
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

// managerNames lists the configured device managers, sorted.
func managerNames(cfg config.Service) []string {
	names := make([]string, 0, len(cfg.DeviceManager))
	for name := range cfg.DeviceManager {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
