package api

import (
	"context"
	"fmt"
	device2 "github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/unity"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func managerForPlatform(t models.PlatformType) string {
	switch t {
	case models.PlatformTypeiOS:
		return "ios_device"
	case models.PlatformTypeAndroid:
		return "android_device"
	case models.PlatformTypeMac:
		return "macos"
	case models.PlatformTypeWindows:
		return "windows"
	case models.PlatformTypeLinux:
		return "linux"
	case models.PlatformTypeWeb:
		return "web"
	case models.PlatformTypeEditor:
		return "unity_editor"
	case models.PlatformTypeiOSSimulator:
		return "iossim"
	}
	return ""
}

func (s *Service) getDevices(c *gin.Context, project *models.Project) {
	p := c.Query("platform")

	var devices []models.Device
	query := s.db
	if p != "" {
		platform, _ := strconv.ParseInt(p, 10, 64)
		query = query.Where("platform_type = ?", models.PlatformType(platform))
	}
	if err := query.Preload("Node").Find(&devices).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	for i := range devices {
		dev, _ := s.devicesManager.GetDevice(devices[i].DeviceIdentifier)
		devices[i].Dev = dev
		if devices[i].NodeID > 0 {
			devices[i].Status = device2.StateRemoteDisconnected
		} else {
			devices[i].Status = device2.StateUnknown
		}

		if dev != nil {
			devices[i].Status = dev.DeviceState()
			if dev.Connection() != nil {
				devices[i].Connection = dev.Connection().ConnectionParameter
			}
		}
	}

	c.JSON(http.StatusOK, devices)
}

func (s *Service) unlockDevice(c *gin.Context, project *models.Project) {
	deviceID := c.Param("device_id")
	_ = deviceID
	var device models.Device
	if err := s.db.Preload("ConnectionParameter").Preload("DeviceParameter").Preload("CustomParameter").Find(&device, "id = ?", deviceID).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	dev, _ := s.devicesManager.GetDevice(device.DeviceIdentifier)
	if dev.IsLocked() {
		err := dev.Unlock()
		if err != nil {
			s.error(c, http.StatusConflict, err)
			return
		}
	}
	c.JSON(http.StatusOK, device)
}

func (s *Service) getDevice(c *gin.Context, project *models.Project) {

	deviceID := c.Param("device_id")
	_ = deviceID
	var device models.Device

	if err := s.db.Preload("ConnectionParameter").Preload("DeviceParameter").Preload("CustomParameter").Find(&device, "id = ?", deviceID).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	dev, _ := s.devicesManager.GetDevice(device.DeviceIdentifier)
	device.Dev = dev
	device.Status = device2.StateUnknown
	if dev != nil {
		device.Status = dev.DeviceState()
		if dev.Connection() != nil {
			device.Connection = dev.Connection().ConnectionParameter
		}
	}

	c.JSON(http.StatusOK, device)
}

func (s *Service) deleteDevice(c *gin.Context, project *models.Project) {
	deviceID := c.Param("device_id")

	var device models.Device
	if err := s.db.First(&device, deviceID).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	if err := s.db.Delete(&device).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) updateDevice(c *gin.Context, project *models.Project) {
	deviceID := c.Param("device_id")

	var dev models.Device
	if err := c.Bind(&dev); err != nil {
		return
	}

	var device models.Device
	if err := s.db.Preload("ConnectionParameter").Preload("CustomParameter").First(&device, deviceID).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	var ids []uint
	for i := range device.CustomParameter {
		exists := false
		for d := range dev.CustomParameter {
			if device.CustomParameter[i].Key == dev.CustomParameter[d].Key {
				exists = true
				break
			}
		}
		if !exists {
			ids = append(ids, device.CustomParameter[i].ID)
		}
	}

	if len(ids) > 0 {
		if err := s.db.Delete(&models.CustomParameter{}, "id in (?)", ids).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
	}

	if err := s.db.Updates(&dev).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

func (s *Service) deviceRunTests(c *gin.Context, project *models.Project) {
	type Request struct {
		TestName string
		Env      string
	}
	type Response struct {
		Success bool
		Message string
	}
	var req Request
	c.Bind(&req)
	arr := strings.Split(req.TestName, " ")
	if len(arr) != 2 {
		s.error(c, http.StatusNotFound, fmt.Errorf("invalid method signature"))
		return
	}

	deviceID := c.Param("device_id")
	_ = deviceID
	var device models.Device
	if err := s.db.Find(&device, "id = ?", deviceID).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	dev, _ := s.devicesManager.GetDevice(device.DeviceIdentifier)
	if dev == nil {
		s.error(c, http.StatusNotFound, fmt.Errorf("real device not found"))
		return
	}

	if err := dev.Lock(); err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}
	defer func() {
		_ = dev.Unlock()
	}()

	if !dev.IsAppConnected() {
		s.error(c, http.StatusNotFound, fmt.Errorf("no app connected"))
		return
	}
	if false {
		var testsAction action.TestsGet
		s.devicesManager.SendAction(dev, &testsAction)
		logrus.Info("tests %v", testsAction)
		time.Sleep(2 * time.Minute)
		for _, test := range testsAction.Tests {
			runTestAction := action.TestStart{
				Assembly: test.Assembly,
				Class:    test.Class,
				Method:   test.Method,
			}
			s.devicesManager.SendAction(dev, &runTestAction)
			time.Sleep(3 * time.Minute)
		}
		return
	}

	reset := action.UnityReset{}
	s.devicesManager.SendAction(dev, &reset)
	envParams := extractParams(req.Env)

	runTestAction := action.TestStart{
		Class:  arr[0],
		Method: arr[1],
		Env:    envParams,
	}

	executor := unity.NewExecutor(s.devicesManager, nil)

	if err := executor.Execute(context.Background(), dev, runTestAction, 5*time.Minute); err != nil {
		logrus.Errorf("Execute failed: %v", err)
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	logrus.Debug("Execution Finished")
	c.JSON(http.StatusOK, &Response{true, "Execution Finished"})
}
