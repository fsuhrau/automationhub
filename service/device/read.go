package device

import (
	"context"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
)

// List returns the project's devices with their live state; a non-nil
// platform limits the list to that platform.
func (s *Service) List(ctx context.Context, project *models.Project, platform *models.PlatformType) ([]models.Device, error) {
	devices, err := s.deps.Repo.List(ctx, project, platform)
	if err != nil {
		return nil, err
	}
	for i := range devices {
		s.withLive(&devices[i])
	}
	return devices, nil
}

// Get returns one device of the project with its parameters and live state.
func (s *Service) Get(ctx context.Context, project *models.Project, id string) (*models.Device, error) {
	d, err := s.deps.Repo.FindDetail(ctx, project, id)
	if err != nil {
		return nil, err
	}
	s.withLive(d)
	return d, nil
}

// Delete removes a device of the project.
func (s *Service) Delete(ctx context.Context, project *models.Project, id string) error {
	d, err := s.deps.Repo.Find(ctx, project, id)
	if err != nil {
		return err
	}
	return s.deps.Repo.Delete(ctx, d)
}

// Unlock ends a live control lease on the device and unlocks it. The device
// has to be connected to the hub.
func (s *Service) Unlock(ctx context.Context, project *models.Project, id string) (*models.Device, error) {
	d, err := s.deps.Repo.FindDetail(ctx, project, id)
	if err != nil {
		return nil, err
	}
	dev, _ := s.deps.Devices.GetDevice(d.DeviceIdentifier)
	if dev == nil {
		return nil, apperr.NotFound("device %s is not connected", d.DeviceIdentifier)
	}
	s.deps.Leases.ForceRelease(dev.DeviceID())
	if dev.IsLocked() {
		if err := events.UnlockDevice(dev); err != nil {
			return nil, apperr.Conflict("%v", err)
		}
	}
	return d, nil
}

// ProtocolQuery asks for the newest Limit results on the device with DeviceID.
type ProtocolQuery struct {
	DeviceID string
	Limit    int
}

// ProtocolScope is ProtocolQuery resolved to a device of a project.
type ProtocolScope struct {
	ProjectID, DeviceID uint
	Limit               int
}

// Protocols lists the latest results of tests of the project that ran on the
// device, newest first.
func (s *Service) Protocols(ctx context.Context, project *models.Project, q ProtocolQuery) ([]models.TestProtocol, error) {
	d, err := s.deps.Repo.Find(ctx, project, q.DeviceID)
	if err != nil {
		return nil, err
	}
	protocols, err := s.deps.Repo.Protocols(ctx, ProtocolScope{ProjectID: project.ID, DeviceID: d.ID, Limit: q.Limit})
	if err != nil {
		return nil, err
	}
	for i := range protocols {
		s.deps.CleanProtocol(&protocols[i])
	}
	return protocols, nil
}

// ControlDevice returns the live device of the project for live control; it
// has to be connected to the hub.
func (s *Service) ControlDevice(ctx context.Context, project *models.Project, id string) (device.Device, error) {
	d, err := s.deps.Repo.Find(ctx, project, id)
	if err != nil {
		return nil, err
	}
	dev, _ := s.deps.Devices.GetDevice(d.DeviceIdentifier)
	if dev == nil {
		return nil, apperr.Conflict("device %s is not connected", d.DeviceIdentifier)
	}
	return dev, nil
}

// ControlApp returns the install parameters of an app bundle of the
// project; an archived bundle has no file to install.
func (s *Service) ControlApp(ctx context.Context, project *models.Project, binaryID uint) (*app.Parameter, error) {
	binary, err := s.deps.Repo.Bundle(ctx, project, binaryID)
	if err != nil {
		return nil, err
	}
	if binary.Archived() {
		return nil, apperr.Invalid("app bundle %d is archived; upload its file again to install it", binaryID)
	}
	params := base.GetParams(binary, "")
	return &params, nil
}
