package config

func (s *Service) GetDeviceConfig(id string) *Device {
	for i := range s.Devices {
		if s.Devices[i].ID == id {
			return &s.Devices[i]
		}
	}
	return nil
}

func (s *Service) GetDevicesForManager(manager string) []Device {
	var devices []Device
	for i := range s.Devices {
		if s.Devices[i].Manager == manager {
			devices = append(devices, s.Devices[i])
		}
	}
	return devices
}