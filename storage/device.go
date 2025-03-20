package storage

import (
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

type Device interface {
	GetDevices(manager string) (models.Devices, error)
	NewDevice(manager string, device models.Device) error
	GetDevice(manager, deviceId string) (*models.Device, error)
	Update(manager string, dev device.Device) error
}

type deviceStore struct {
	devices map[string]models.Devices
	db      *gorm.DB
}

func NewDeviceStore(db *gorm.DB) *deviceStore {
	store := &deviceStore{
		db:      db,
		devices: make(map[string]models.Devices),
	}

	return store
}

func (d *deviceStore) GetDevices(manager string) (models.Devices, error) {
	if devices, ok := d.devices[manager]; ok {
		return devices, nil
	}

	var devices models.Devices
	if err := d.db.Where("manager = ?", manager).Preload("Parameter").Preload("ConnectionParameter").Find(&devices).Error; err != nil {
		return nil, err
	}

	d.devices[manager] = devices
	return devices, nil
}

func (d *deviceStore) GetDevice(manager, deviceId string) (*models.Device, error) {
	devices, err := d.GetDevices(manager)
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
	tx := d.db.Begin()

	_, err := d.GetDevice(manager, dev.DeviceIdentifier)
	if err == nil {
		tx.Rollback()
		return fmt.Errorf("device with the same id exists already")
	}

	if err := d.db.Create(&dev).Error; err != nil {
		tx.Rollback()
		return err
	}

	if dev.ConnectionParameter != nil {
		dev.ConnectionParameter.DeviceID = dev.ID
		if err := d.db.FirstOrCreate(&dev.ConnectionParameter).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	d.devices[manager] = append(d.devices[manager], &dev)
	tx.Commit()
	return nil
}

func (d *deviceStore) Update(manager string, dev device.Device) error {
	tx := db.Begin()
	deviceData, err := d.GetDevice(manager, dev.DeviceID())
	if err != nil {
		return err
	}

	needsUpdate := false
	if deviceData.Name != dev.DeviceName() {
		deviceData.Name = dev.DeviceName()
		needsUpdate = true
	}
	if deviceData.OS != dev.DeviceOSName() {
		deviceData.OS = dev.DeviceOSName()
		needsUpdate = true
	}
	if deviceData.OSVersion != dev.DeviceOSVersion() {
		deviceData.OSVersion = dev.DeviceOSVersion()
		needsUpdate = true
	}

	if deviceData.HardwareModel != dev.DeviceModel() {
		deviceData.HardwareModel = dev.DeviceModel()
		needsUpdate = true
	}

	if deviceData.Manager != manager {
		deviceData.Manager = manager
		needsUpdate = true
	}

	statusUpdate := false
	if deviceData.Status != dev.DeviceState() {
		deviceData.Status = dev.DeviceState()
		statusUpdate = true
	}

	if needsUpdate || statusUpdate {
		d.db.Updates(deviceData)
	}

	if statusUpdate {
		log := models.DeviceLog{
			DeviceID: deviceData.ID,
			Status:   dev.DeviceState(),
			Payload:  "",
		}
		d.db.Create(&log)
		events.DeviceStatusChanged.Trigger(events.DeviceStatusChangedPayload{
			DeviceID:    deviceData.ID,
			DeviceState: uint(dev.DeviceState()),
		})
	}
	tx.Commit()
	return nil
}
