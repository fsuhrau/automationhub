package iosdevice

import (
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
)

// attached is an entry of the usbmuxd device list with what lockdown said.
type attached struct {
	serial string
	values *ios.GetAllValuesResponse // nil when lockdown refused
	err    error
}

// RefreshDevices reads the attached devices from usbmuxd and lockdown. While
// the usbmuxd event subscription is up, unforced calls only run as a safety
// net every fallbackPollInterval.
func (m *Handler) RefreshDevices(force bool) error {
	if !m.dueForPoll(force) {
		return nil
	}
	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()
	found, err := m.readAttached()
	if err != nil {
		return err
	}
	m.applyAttached(found)
	return nil
}

func (m *Handler) dueForPoll(force bool) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.deps.Clock.Now()
	if !force && m.tracking.Load() && now.Sub(m.lastPoll) < fallbackPollInterval {
		return false
	}
	m.lastPoll = now
	return true
}

// readAttached lists the devices and reads their values (lockdown calls).
func (m *Handler) readAttached() ([]attached, error) {
	list, err := m.deps.USB.ListDevices()
	if err != nil {
		return nil, err
	}
	found := make([]attached, 0, len(list.DeviceList))
	for _, entry := range list.DeviceList {
		a := attached{serial: entry.Properties.SerialNumber}
		if values, err := m.deps.USB.GetValues(entry); err != nil {
			a.err = err
		} else {
			a.values = &values
		}
		found = append(found, a)
	}
	return found, nil
}

func (m *Handler) applyAttached(found []attached) {
	seen := m.deps.Clock.Now().UTC()
	for _, a := range found {
		m.applyOne(a, seen)
	}
	// known devices that aren't attached (unplugged, network devices away)
	for _, dev := range m.List() {
		if !dev.lastUpdateAt.Equal(seen) {
			dev.State = device.StateRemoteDisconnected
			_ = m.deviceStorage.Update(m.Name(), dev)
		}
	}
}

func (m *Handler) applyOne(a attached, seen time.Time) {
	dev, known := m.Get(a.serial)
	if !known {
		dev = m.createDevice(a)
	}
	dev.apply(a, seen)
	if known {
		_ = m.deviceStorage.Update(m.Name(), dev)
		return
	}
	m.Add(dev)
	m.registerNewDevice(dev)
}

// createDevice makes a device for an entry; one that lockdown refuses only has
// its serial to go by.
func (m *Handler) createDevice(a attached) *Device {
	dev := m.newDevice(a.serial)
	if a.values == nil {
		dev.Name, dev.OSName = a.serial, "iOS"
	}
	return dev
}

// apply takes over what the list and lockdown said at seen.
func (d *Device) apply(a attached, seen time.Time) {
	d.lastUpdateAt = seen
	if a.err != nil {
		// attached but lockdown refuses: not trusted yet, or locked with a passcode
		logrus.Debugf("ios device %s not readable: %v", a.serial, a.err)
		d.State = device.StateUnauthorized
		return
	}
	d.UpdateDeviceInfos(*a.values)
	// attached and readable; WebDriverAgent is started on demand
	// (StartDevice), so it doesn't decide whether the device is ready
	d.State = device.StateBooted
}

// registerNewDevice stores a newly seen device and wires its remote logging.
func (m *Handler) registerNewDevice(d *Device) {
	dev := models.Device{
		DeviceIdentifier: d.ID,
		DeviceType:       models.DeviceTypePhone,
		Name:             d.Name,
		Manager:          Manager,
		PlatformType:     models.PlatformTypeiOS,
		OS:               d.OSName,
		OSVersion:        d.OSVersion,
		ConnectionParameter: &models.ConnectionParameter{
			ConnectionType: models.ConnectionTypeUSB,
		},
	}
	m.wireRemote(d)
	_ = m.deviceStorage.NewDevice(m.Name(), dev)
	_ = m.deviceStorage.Update(m.Name(), d)
	d.SetConfig(&dev)
}
