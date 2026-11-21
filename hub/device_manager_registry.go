package hub

import (
	"fmt"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

func (dm *DeviceManager) Managers() []string {
	var manager []string
	for k := range dm.deviceHandlers {
		manager = append(manager, k)
	}
	return manager
}

func (dm *DeviceManager) AddHandler(manager device.Handler) {
	dm.deviceHandlers[manager.Name()] = manager
}

func (dm *DeviceManager) ListDevices() {
	for _, m := range dm.deviceHandlers {
		dm.log.Infof("Manager: %s", m.Name())
		devices, err := m.GetDevices()
		if err != nil {
			dm.log.Errorf("failed: %v", err)
		}
		for _, d := range devices {
			dm.log.Infof("Device %s (%s) %d %s", d.DeviceName(), d.DeviceID(), d.DeviceState(), d.DeviceOSName())
		}
	}
}

func (dm *DeviceManager) Devices() (map[string][]device.Device, error) {
	devices := make(map[string][]device.Device)
	for h, m := range dm.deviceHandlers {
		d, err := m.GetDevices()
		if err != nil {
			dm.log.Errorf("failed: %v", err)
		}
		devices[h] = d
	}
	return devices, nil
}

func (dm *DeviceManager) Start(dev device.Device) error {
	for _, m := range dm.deviceHandlers {
		if m.HasDevice(dev) {
			defer dm.notifier.Notify()
			return m.StartDevice(dev.DeviceID())
		}
	}
	return device.DeviceNotFoundError
}

func (dm *DeviceManager) Stop(dev device.Device) error {
	for _, m := range dm.deviceHandlers {
		if m.HasDevice(dev) {
			defer dm.notifier.Notify()
			return m.StopDevice(dev.DeviceID())
		}
	}
	return device.DeviceNotFoundError
}

func (dm *DeviceManager) FindDevice(params *app.Parameter) device.Device {
	for _, mng := range dm.deviceHandlers {
		devices, _ := mng.GetDevices()
		for i := range devices {
			if devices[i].DeviceOSName() == params.Platform {
				return devices[i]
			}
		}
	}
	return nil
}

func (dm *DeviceManager) RegisterDevice(data device.RegisterData) (device.Device, error) {
	for _, mng := range dm.deviceHandlers {
		if mng.Name() == data.ManagerType {
			return mng.RegisterDevice(data)
		}
	}
	return nil, fmt.Errorf("device for type not found")
}

func (dm *DeviceManager) GetDevice(id string) (device.Device, string) {
	for manager, mng := range dm.deviceHandlers {
		devices, _ := mng.GetDevices()
		for i := range devices {
			if devices[i].DeviceID() == id {
				return devices[i], manager
			}
		}
	}

	return nil, ""
}
