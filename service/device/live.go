package device

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

// withLive sets the state a device has right now on its stored record:
// the status, lock, who controls it, and the live device and connection.
// A device the pool doesn't know is unknown, or disconnected from its node
// when it belongs to one. This is the only place that merges the pool and
// the leases into the stored data.
func (s *Service) withLive(d *models.Device) {
	dev, _ := s.deps.Devices.GetDevice(d.DeviceIdentifier)
	d.Dev = dev
	if dev == nil {
		d.Status = offlineState(d)
		return
	}
	d.Status = dev.DeviceState()
	d.IsLocked = dev.IsLocked()
	if lease, ok := s.deps.Leases.Holder(dev.DeviceID()); ok {
		d.LockedBy = lease.Owner
	}
	if conn := dev.Connection(); conn != nil {
		d.Connection = conn.ConnectionParameter
	}
}

func offlineState(d *models.Device) device.State {
	if d.NodeID > 0 {
		return device.StateNodeDisconnected
	}
	return device.StateUnknown
}
