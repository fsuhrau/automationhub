package iossim

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/storage/models"
)

// RefreshDevices reads the simulators from simctl, stores new and changed ones
// and marks the ones simctl no longer reports.
func (m *Handler) RefreshDevices(bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	sims, err := m.simctl.list()
	if err != nil {
		return err
	}
	now := m.deps.Clock.Now().UTC()
	seen := make(map[string]bool, len(sims))
	for _, sim := range sims {
		seen[sim.UDID] = true
		m.apply(sim, now)
	}
	for _, d := range m.List() {
		if !seen[d.ID] {
			d.State = missingState(d)
			m.store(d)
		}
	}
	return nil
}

// apply updates the known simulator or adds and announces a new one.
func (m *Handler) apply(sim simInfo, now time.Time) {
	d, known := m.Get(sim.UDID)
	if !known {
		d = NewDevice(generic.Info{ID: sim.UDID}, m.deps)
		m.Add(d)
		m.announce(sim)
	}
	d.Name, d.OSName, d.OSVersion, d.lastUpdateAt = sim.Name, sim.OSName, sim.OSVersion, now
	d.State = sim.State
	m.store(d)
}

func (m *Handler) announce(sim simInfo) {
	err := m.deviceStorage.NewDevice(Manager, models.Device{
		DeviceIdentifier:    sim.UDID,
		DeviceType:          models.DeviceTypePhone,
		Name:                sim.Name,
		Manager:             Manager,
		PlatformType:        models.PlatformTypeiOSSimulator,
		OS:                  sim.OSName,
		OSVersion:           sim.OSVersion,
		ConnectionParameter: &models.ConnectionParameter{ConnectionType: models.ConnectionTypeUSB},
	})
	if err != nil {
		logrus.Warnf("ios sim: store new simulator %s: %v", sim.UDID, err)
	}
}

func (m *Handler) store(d *Device) {
	if err := m.deviceStorage.Update(Manager, d); err != nil {
		logrus.Warnf("ios sim: update simulator %s: %v", d.ID, err)
	}
}

// missingState is the state of a simulator simctl does not report: remote
// ones are disconnected, local ones unknown.
func missingState(d *Device) device.State {
	cfg := d.GetConfig()
	if cfg != nil && cfg.ConnectionParameter != nil && cfg.ConnectionParameter.ConnectionType == models.ConnectionTypeRemote {
		return device.StateRemoteDisconnected
	}
	return device.StateUnknown
}

var _ device.Handler = (*Handler)(nil)
