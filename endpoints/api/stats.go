package api

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) getStats(c *gin.Context) {
	type Response struct {
		AppsCount          int
		AppsStorageSize    int
		DatabaseSize       int
		SystemMemoryUsage  int
		SystemUptime       int
		TestsLastProtocols []models.TestProtocol
		TestsLastFailed    []models.TestProtocol
		DeviceCount        int
		DeviceBooted       int
	}

	var response Response

	response.AppsCount, response.AppsStorageSize = s.getAppStats()

	response.DeviceCount, response.DeviceBooted = s.getDeviceStats()

	response.TestsLastProtocols, response.TestsLastFailed = s.getProtocolStats()

	c.JSON(http.StatusOK, response)
}

func (s *Service) getAppStats() (int, int) {
	var apps []models.App
	if err := s.db.Find(&apps).Error; err != nil {
		return 0, 0
	}

	count := len(apps)
	size := 0
	for i := range apps {
		size += apps[i].Size
	}

	return count, size
}

func (s *Service) getDeviceStats() (int, int) {
	var devices []models.Device
	if err := s.db.Find(&devices).Error; err != nil {
		return 0, 0
	}

	count := len(devices)
	booted := 0
	for i := range devices {
		dev := s.devicesManager.GetDevice(devices[i].DeviceIdentifier)
		if dev != nil && dev.DeviceState() >= device.StateBooted {
			booted++
		}
	}

	return count, booted
}

func (s *Service) getProtocolStats() ([]models.TestProtocol, []models.TestProtocol) {
	var protocols []models.TestProtocol
	var failedProtocols []models.TestProtocol
	s.db.Preload("Device").Order("created_at desc").Limit(10).Find(&protocols)

	s.db.Preload("Device").Where("test_result = ?", models.TestResultFailed).Order("created_at desc").Limit(10).Find(&failedProtocols)

	return protocols, failedProtocols
}
