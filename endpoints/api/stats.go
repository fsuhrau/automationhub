package api

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) getStats(c *gin.Context) {
	type Response struct {
		AppsCount          int                   `json:"appsCount"`
		AppsStorageSize    int                   `json:"appsStorageSize"`
		DatabaseSize       int                   `json:"databaseSize"`
		SystemMemoryUsage  int                   `json:"systemMemoryUsage"`
		SystemUptime       int                   `json:"systemUptime"`
		TestsLastProtocols []models.TestProtocol `json:"testsLastProtocols"`
		TestsLastFailed    []models.TestProtocol `json:"testsLastFailed"`
		DeviceCount        int                   `json:"deviceCount"`
		DeviceBooted       int                   `json:"deviceBooted"`
	}

	var response Response

	response.AppsCount, response.AppsStorageSize = s.getAppStats()

	response.DeviceCount, response.DeviceBooted = s.getDeviceStats()

	response.TestsLastProtocols, response.TestsLastFailed = s.getProtocolStats()

	c.JSON(http.StatusOK, response)
}

func (s *Service) getAppStats() (int, int) {
	var apps []models.AppBinary
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
		dev, _ := s.devicesManager.GetDevice(devices[i].DeviceIdentifier)
		if dev != nil && dev.DeviceState() >= device.StateBooted {
			booted++
		}
	}

	return count, booted
}

func (s *Service) getProtocolStats() ([]models.TestProtocol, []models.TestProtocol) {
	var protocols []models.TestProtocol
	var failedProtocols []models.TestProtocol
	s.db.Preload("TestRun").Preload("TestRun.Test").Preload("Device").Order("created_at desc").Limit(10).Find(&protocols)
	s.db.Preload("TestRun").Preload("TestRun.Test").Preload("Device").Where("test_result = ?", models.TestResultFailed).Order("created_at desc").Limit(10).Find(&failedProtocols)

	return protocols, failedProtocols
}
