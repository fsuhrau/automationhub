package storage

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// Update stores what differs between the live device and its record and
// notifies about it.
func (d *deviceStore) Update(manager string, dev device.Device) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := d.getDevice(manager, dev.DeviceID())
	if err != nil {
		return err
	}
	needsUpdate := syncDeviceInfo(data, manager, dev)
	statusUpdate := data.Status != dev.DeviceState()
	if !needsUpdate && !statusUpdate {
		return nil
	}
	data.Status = dev.DeviceState()
	if err := d.saveDevice(data, statusUpdate); err != nil {
		return err
	}
	if needsUpdate {
		d.notify.DeviceListChanged(events.DeviceListChangedPayload{DeviceID: data.ID})
	}
	if statusUpdate {
		d.notify.DeviceStatusChanged(events.DeviceStatusChangedPayload{
			DeviceID:    data.ID,
			DeviceState: uint(dev.DeviceState()),
		})
	}
	return nil
}

// syncDeviceInfo copies name, OS, manager and parameters of dev into data and
// reports whether anything changed.
func syncDeviceInfo(data *models.Device, manager string, dev device.Device) bool {
	changed := false
	setString := func(target *string, value string) {
		if *target != value {
			*target = value
			changed = true
		}
	}
	setString(&data.Name, dev.DeviceName())
	setString(&data.OS, dev.DeviceOSName())
	setString(&data.OSVersion, dev.DeviceOSVersion())
	setString(&data.Manager, manager)
	if syncDeviceParameters(data, dev.DeviceParameter()) {
		changed = true
	}
	return changed
}

// syncDeviceParameters sets or adds every parameter and reports a change.
func syncDeviceParameters(data *models.Device, parameter map[string]string) bool {
	changed := false
	for key, value := range parameter {
		if setDeviceParameter(data, key, value) {
			changed = true
		}
	}
	return changed
}

func setDeviceParameter(data *models.Device, key, value string) bool {
	for i := range data.DeviceParameter {
		if data.DeviceParameter[i].Key != key {
			continue
		}
		if data.DeviceParameter[i].Value == value {
			return false
		}
		data.DeviceParameter[i].Value = value
		return true
	}
	data.DeviceParameter = append(data.DeviceParameter, models.DeviceParameter{Key: key, Value: value})
	return true
}

// saveDevice writes the device and, for a status change, its log entry.
func (d *deviceStore) saveDevice(data *models.Device, statusUpdate bool) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(data).Error; err != nil {
			return err
		}
		if !statusUpdate {
			return nil
		}
		return tx.Create(&models.DeviceLog{DeviceID: data.ID, Status: data.Status}).Error
	})
}
