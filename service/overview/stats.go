package overview

import (
	"context"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

// protocolsListed is how many of the newest and of the newest failed
// protocols the statistics list.
const protocolsListed = 10

// Stats are the statistics of one project. The database size and the
// system figures have no source any more and stay 0; the fields are part of
// the api.
type Stats struct {
	AppsCount          int                   `json:"appsCount"`
	AppsStorageSize    int                   `json:"appsStorageSize"`
	DatabaseSize       int                   `json:"databaseSize"`
	SystemMemoryUsage  int                   `json:"systemMemoryUsage"`
	SystemUptime       int                   `json:"systemUptime"`
	TestsLastProtocols []models.TestProtocol `json:"testsLastProtocols"`
	TestsLastFailed    []models.TestProtocol `json:"testsLastFailed"`
	DeviceCount        int                   `json:"deviceCount"`
	DeviceBooted       int                   `json:"deviceBooted"`
}

// Stats counts the project's bundles and visible devices and lists its
// newest and newest failed protocols.
func (s *Service) Stats(ctx context.Context, project *models.Project) (*Stats, error) {
	totals, err := s.deps.Repo.Totals(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	stats := &Stats{AppsCount: int(totals.Bundles), AppsStorageSize: int(totals.BundleStorage)}
	if stats.TestsLastProtocols, err = s.protocols(ctx, project.ID, false); err != nil {
		return nil, err
	}
	if stats.TestsLastFailed, err = s.protocols(ctx, project.ID, true); err != nil {
		return nil, err
	}
	devices, err := s.deps.Repo.Devices(ctx, project)
	if err != nil {
		return nil, err
	}
	stats.DeviceCount, stats.DeviceBooted = len(devices), s.booted(devices)
	return stats, nil
}

func (s *Service) protocols(ctx context.Context, projectID uint, failedOnly bool) ([]models.TestProtocol, error) {
	return s.deps.Repo.Protocols(ctx, ProtocolQuery{ProjectID: projectID, FailedOnly: failedOnly, Limit: protocolsListed})
}

// booted counts the devices the pool has live and booted.
func (s *Service) booted(devices []models.Device) int {
	booted := 0
	for i := range devices {
		dev, _ := s.deps.Pool.GetDevice(devices[i].DeviceIdentifier)
		if dev != nil && dev.DeviceState() >= device.StateBooted {
			booted++
		}
	}
	return booted
}
