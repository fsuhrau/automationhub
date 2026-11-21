package api

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
)

// onDeviceControlChanged tells the device pool to reload when a device got a
// new live controller or lost it, so "controlled by" shows up right away.
func (s *Service) onDeviceControlChanged(deviceIdentifier string) {
	var device models.Device
	if err := s.db.Select("id").First(&device, "device_identifier = ?", deviceIdentifier).Error; err != nil {
		s.logger.Debugf("control change for unknown device %s: %v", deviceIdentifier, err)
		return
	}
	events.DeviceListChanged.Trigger(events.DeviceListChangedPayload{DeviceID: device.ID})
}
