package androiddevice

import "github.com/fsuhrau/automationhub/device"

// RefreshDevices polls "adb devices -l". While the adb event stream is
// connected this only runs as a safety net every fallbackPollInterval.
func (m *Handler) RefreshDevices(force bool) error {
	if !m.shouldPoll(force) {
		return nil
	}
	output, err := m.adbHost("devices", "-l")
	if err != nil {
		return err
	}
	m.applyDeviceList(parseADBDeviceList(string(output)))
	return nil
}

// shouldPoll reports whether the list must be polled now and, if so, records
// the poll time.
func (m *Handler) shouldPoll(force bool) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.deps.Clock.Now()
	if !force && m.tracking.Load() && now.Sub(m.lastPoll) < fallbackPollInterval {
		return false
	}
	m.lastPoll = now
	return true
}

// applyDeviceList updates all devices from a complete adb device list.
func (m *Handler) applyDeviceList(entries []adbDeviceEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	listed := make(map[string]bool, len(entries))
	for _, entry := range entries {
		listed[entry.Serial] = true
		m.applyEntry(entry)
	}
	m.markUnlisted(listed)
}

func (m *Handler) applyEntry(entry adbDeviceEntry) {
	if entry.State != adbStateDevice {
		m.updateUnavailableDevice(entry)
		return
	}
	if d, ok := m.Get(entry.Serial); ok {
		m.refreshKnown(d, entry)
		return
	}
	m.addNew(entry)
}

func (m *Handler) refreshKnown(d *Device, entry adbDeviceEntry) {
	// reads the device properties on first use (e.g. after authorization)
	d.UpdateDeviceInfos(entry)
	d.State = device.StateBooted
	m.persist(d)
}

func (m *Handler) addNew(entry adbDeviceEntry) {
	d := newDevice(m.deps, m.video, entry.Serial)
	d.UpdateDeviceInfos(entry)
	d.State = device.StateBooted
	m.Add(d)
	m.registerNewDevice(d)
}

// updateUnavailableDevice handles adb entries that are listed but not usable,
// e.g. "unauthorized" (USB debugging not accepted on the phone) or "offline".
func (m *Handler) updateUnavailableDevice(entry adbDeviceEntry) {
	state := device.StateRemoteDisconnected
	if entry.State == adbStateUnauthorized {
		state = device.StateUnauthorized
	}
	if d, ok := m.Get(entry.Serial); ok {
		d.State = state
		m.persist(d)
		return
	}
	if state != device.StateUnauthorized {
		return
	}
	// a new phone waiting for authorization: list it, but its properties can
	// only be read once it is authorized (Parameter stays nil until then)
	d := newDevice(m.deps, m.video, entry.Serial)
	d.State = state
	m.Add(d)
	m.registerNewDevice(d)
}

// markUnlisted marks the known devices adb doesn't list: unplugged, emulator
// stopped, network device not connected.
func (m *Handler) markUnlisted(listed map[string]bool) {
	for _, d := range m.List() {
		if !listed[d.ID] {
			d.State = device.StateRemoteDisconnected
			m.persist(d)
		}
	}
}
