package config

func (m Manager) GetDeviceConfig(id string) *Device {
	for i := range m.Devices {
		if m.Devices[i].ID == id {
			return &m.Devices[i]
		}
	}
	return nil
}

func (m *Manager) GetDevicesForManager(manager string) []Device {
	var devices []Device
	for i := range m.Devices {
		devices = append(devices, m.Devices[i])
	}
	return devices
}
