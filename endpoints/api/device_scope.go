package api

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/devicestore"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// projectDevices scopes device queries to the project's company. The runs
// and test endpoints still query devices themselves; they drop this when
// they move to services.
func (s *Service) projectDevices(project *models.Project) *gorm.DB {
	return devicestore.VisibleTo(s.db, project)
}

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
