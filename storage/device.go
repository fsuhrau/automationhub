package storage

import (
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"sync"
)

type Device interface {
	GetDevices(manager string) (models.Devices, error)
	NewDevice(manager string, device models.Device) error
	GetDevice(manager, deviceId string) (*models.Device, error)
	Update(manager string, dev device.Device) error
}

type deviceStore struct {
	mu      sync.Mutex // guards devices cache and serializes writes
	devices map[string]models.Devices
	db      *gorm.DB
	notify  DeviceNotifier
}

// NewDeviceStore returns a store that announces changes on the event bus.
func NewDeviceStore(db *gorm.DB) *deviceStore {
	return NewDeviceStoreWithNotifier(db, busNotifier{})
}

// NewDeviceStoreWithNotifier returns a store that announces changes to n.
func NewDeviceStoreWithNotifier(db *gorm.DB, n DeviceNotifier) *deviceStore {
	return &deviceStore{
		db:      db,
		devices: make(map[string]models.Devices),
		notify:  n,
	}
}

func (d *deviceStore) GetDevices(manager string) (models.Devices, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.getDevices(manager)
}

func (d *deviceStore) getDevices(manager string) (models.Devices, error) {
	if devices, ok := d.devices[manager]; ok {
		return devices, nil
	}

	var devices models.Devices
	if err := d.db.Where("manager = ?", manager).Preload("DeviceParameter").Preload("CustomParameter").Preload("ConnectionParameter").Find(&devices).Error; err != nil {
		return nil, err
	}

	d.devices[manager] = devices
	return devices, nil
}

func (d *deviceStore) GetDevice(manager, deviceId string) (*models.Device, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.getDevice(manager, deviceId)
}

func (d *deviceStore) getDevice(manager, deviceId string) (*models.Device, error) {
	devices, err := d.getDevices(manager)
	if err != nil {
		return nil, err
	}

	for _, dev := range devices {
		if dev.DeviceIdentifier == deviceId {
			return dev, nil
		}
	}
	return nil, fmt.Errorf("device with id: %s not found", deviceId)
}

func (d *deviceStore) NewDevice(manager string, dev models.Device) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, err := d.getDevice(manager, dev.DeviceIdentifier); err == nil {
		return fmt.Errorf("device with the same id exists already")
	}

	if err := d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&dev).Error; err != nil {
			return err
		}
		if dev.ConnectionParameter != nil {
			dev.ConnectionParameter.DeviceID = dev.ID
			if err := tx.FirstOrCreate(&dev.ConnectionParameter).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	d.devices[manager] = append(d.devices[manager], &dev)
	d.notify.DeviceListChanged(events.DeviceListChangedPayload{DeviceID: dev.ID, Added: true})
	return nil
}
