package api

import (
	"encoding/json"
	"fmt"
	device2 "github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/unity"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net"
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
	}
	return ""
}

func (s *Service) getDevices(c *gin.Context, project *models.Project) {
	p := c.Query("platform")

	var devices []models.Device
	query := s.db
	if p != "" {
		platform, _ := strconv.ParseInt(p, 10, 64)
		query = query.Where("manager = ?", managerForPlatform(models.PlatformType(platform)))
	}
	if err := query.Find(&devices).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	for i := range devices {
		dev := s.devicesManager.GetDevice(devices[i].DeviceIdentifier)
		devices[i].Dev = dev
		devices[i].Status = device2.StateUnknown
		if dev != nil {
			devices[i].Status = dev.DeviceState()
			if dev.Connection() != nil {
				devices[i].Connection = dev.Connection().ConnectionParameter
			}
		}
	}

	c.JSON(http.StatusOK, devices)
}

func (s *Service) registerDevices(msg []byte, conn *websocket.Conn, c *gin.Context) {

	type Request struct {
		Type     string
		DeviceID string
		IP       string
		Version  string
		OS       string
		Name     string
	}

	clientIp := net.ParseIP(c.ClientIP())

	var req Request

	json.Unmarshal(msg, &req)

	register := device2.RegisterData{
		DeviceOSVersion: req.Version,
		Name:            req.Name,
		DeviceOS:        req.OS,
		DeviceID:        req.DeviceID,
		ManagerType:     req.Type,
		DeviceIP:        clientIp,
		Conn:            conn,
	}

	s.devicesManager.RegisterDevice(register)
}

func (s *Service) getDevice(c *gin.Context, project *models.Project) {

	deviceID := c.Param("device_id")
	_ = deviceID
	var device models.Device

	if err := s.db.Preload("ConnectionParameter").Preload("Parameter").Find(&device, "id = ?", deviceID).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	dev := s.devicesManager.GetDevice(device.DeviceIdentifier)
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
	c.Bind(&dev)

	var device models.Device
	if err := s.db.Preload("ConnectionParameter").Preload("Parameter").First(&device, deviceID).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	var ids []uint
	for i := range device.Parameter {
		exists := false
		for d := range dev.Parameter {
			if device.Parameter[i].Key == dev.Parameter[d].Key {
				exists = true
				break
			}
		}
		if !exists {
			ids = append(ids, device.Parameter[i].ID)
		}
	}

	if len(ids) > 0 {
		if err := s.db.Delete(&models.DeviceParameter{}, "id in (?)", ids).Error; err != nil {
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

	dev := s.devicesManager.GetDevice(device.DeviceIdentifier)
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

	executor := unity.NewExecutor(s.devicesManager)
	if err := executor.Execute(dev, runTestAction, 5*time.Minute); err != nil {
		logrus.Errorf("Execute failed: %v", err)
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	logrus.Debug("Execution Finished")
	c.JSON(http.StatusOK, &Response{true, "Execution Finished"})
}
