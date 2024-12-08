package node

import (
	"fmt"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

var (
	ErrManagerNotFound = fmt.Errorf("manager not found")
	ErrDeviceNotFound  = fmt.Errorf("device not found")
)

type deviceStore struct {
	devices map[string]models.Devices
}

func NewMemoryDeviceStore(cfg map[string]config.Manager) *deviceStore {
	var devices map[string]models.Devices
	devices = make(map[string]models.Devices)
	for k, v := range cfg {
		var devs models.Devices

		for i := range v.Devices {
			var params []models.DeviceParameter
			for k, v := range v.Devices[i].Parameter {
				params = append(params, models.DeviceParameter{
					Key:   k,
					Value: v,
				})
			}
			devs = append(devs, &models.Device{
				Manager:          k,
				DeviceIdentifier: v.Devices[i].Identifier,
				Parameter:        params,
			})
		}
		devices[k] = devs
	}
	return &deviceStore{
		devices: devices,
	}
}

func (ds *deviceStore) GetDevices(manager string) (models.Devices, error) {
	if devs, ok := ds.devices[manager]; ok {
		return devs, nil
	}
	return nil, ErrManagerNotFound
}

func (ds *deviceStore) NewDevice(manager string, dev models.Device) error {
	var devs models.Devices
	devs, _ = ds.devices[manager]
	found := false
	for i := range devs {
		if devs[i].DeviceIdentifier == dev.DeviceIdentifier {
			devs[i] = &dev
			found = true
			break
		}
	}
	if !found {
		devs = append(devs, &dev)
	}
	ds.devices[manager] = devs
	return nil
}

func (ds *deviceStore) GetDevice(manager, deviceId string) (*models.Device, error) {
	if devs, ok := ds.devices[manager]; ok {
		for i := range devs {
			if devs[i].DeviceIdentifier == deviceId {
				return devs[i], nil
			}
		}
		return nil, ErrDeviceNotFound
	}

	return nil, ErrManagerNotFound
}

func (ds *deviceStore) Update(manager string, dev device.Device) error {
	return nil
}
