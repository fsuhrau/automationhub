package api

import (
	"context"
	"errors"
	"fmt"
	device2 "github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/unity"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// projectDevices scopes device queries to the project's company. Devices that
// are not assigned to a company (the default for auto-discovered devices) are
// visible to every project.
func (s *Service) projectDevices(project *models.Project) *gorm.DB {
	return s.db.Where("company_id IN ?", []uint{0, project.CompanyID})
}

// findProjectDevice loads the device from the :device_id path parameter and
// writes a 404 when it does not exist or is not visible to the project.
func (s *Service) findProjectDevice(c *gin.Context, project *models.Project, preloads ...string) (*models.Device, bool) {
	query := s.projectDevices(project)
	for _, p := range preloads {
		query = query.Preload(p)
	}
	var device models.Device
	if err := query.First(&device, "id = ?", c.Param("device_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.error(c, http.StatusNotFound, fmt.Errorf("device %s not found", c.Param("device_id")))
		} else {
			s.error(c, http.StatusInternalServerError, err)
		}
		return nil, false
	}
	return &device, true
}

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

func (s *Service) getDevices(c *gin.Context, project *models.Project) {
	p := c.Query("platform")

	var devices []models.Device
	query := s.projectDevices(project)
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

		if dev != nil {
			devices[i].Status = dev.DeviceState()
			devices[i].IsLocked = dev.IsLocked()
			if l, ok := s.control.Holder(dev.DeviceID()); ok {
				devices[i].LockedBy = l.Owner
			}
			if dev.Connection() != nil {
				devices[i].Connection = dev.Connection().ConnectionParameter
			}
		} else {
			if devices[i].NodeID > 0 {
				devices[i].Status = device2.StateNodeDisconnected
			} else {
				devices[i].Status = device2.StateUnknown
			}
		}
	}

	c.JSON(http.StatusOK, devices)
}

func (s *Service) unlockDevice(c *gin.Context, project *models.Project) {
	device, ok := s.findProjectDevice(c, project, "ConnectionParameter", "DeviceParameter", "CustomParameter")
	if !ok {
		return
	}

	dev, _ := s.devicesManager.GetDevice(device.DeviceIdentifier)
	if dev == nil {
		s.error(c, http.StatusNotFound, fmt.Errorf("device %s is not connected", device.DeviceIdentifier))
		return
	}
	// an explicit unlock also ends a live control lease
	s.control.ForceRelease(dev.DeviceID())
	if dev.IsLocked() {
		err := events.UnlockDevice(dev)
		if err != nil {
			s.error(c, http.StatusConflict, err)
			return
		}
	}
	c.JSON(http.StatusOK, device)
}

func (s *Service) getDevice(c *gin.Context, project *models.Project) {
	device, ok := s.findProjectDevice(c, project, "ConnectionParameter", "DeviceParameter", "CustomParameter")
	if !ok {
		return
	}

	dev, _ := s.devicesManager.GetDevice(device.DeviceIdentifier)
	device.Dev = dev
	device.Status = device2.StateUnknown
	if dev != nil {
		device.Status = dev.DeviceState()
		device.IsLocked = dev.IsLocked()
		if l, ok := s.control.Holder(dev.DeviceID()); ok {
			device.LockedBy = l.Owner
		}
		if dev.Connection() != nil {
			device.Connection = dev.Connection().ConnectionParameter
		}
	}

	c.JSON(http.StatusOK, device)
}

func (s *Service) deleteDevice(c *gin.Context, project *models.Project) {
	device, ok := s.findProjectDevice(c, project)
	if !ok {
		return
	}
	if err := s.db.Delete(device).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// deviceUpdate is what the device edit page changes: the alias and the
// custom parameters. Everything else is reported by the device or its node.
type deviceUpdate struct {
	Alias           string `json:"alias"`
	CustomParameter []struct {
		ID    uint   `json:"id"`
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"customParameter"`
}

// updateDevice sets the alias (an empty one clears it) and replaces the
// custom parameters: submitted ones with the id of an existing parameter are
// updated, others are added, and existing ones not submitted are removed.
// Keys must be set and unique.
func (s *Service) updateDevice(c *gin.Context, project *models.Project) {
	var req deviceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	seen := map[string]bool{}
	for i := range req.CustomParameter {
		key := strings.TrimSpace(req.CustomParameter[i].Key)
		if key == "" {
			s.error(c, http.StatusBadRequest, fmt.Errorf("custom parameter %d has no key", i+1))
			return
		}
		if seen[key] {
			s.error(c, http.StatusBadRequest, fmt.Errorf("custom parameter %q is there twice", key))
			return
		}
		seen[key] = true
		req.CustomParameter[i].Key = key
	}

	device, ok := s.findProjectDevice(c, project, "CustomParameter")
	if !ok {
		return
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Device{}).Where("id = ?", device.ID).Update("alias", strings.TrimSpace(req.Alias)).Error; err != nil {
			return err
		}
		existing := map[uint]bool{}
		for _, p := range device.CustomParameter {
			existing[p.ID] = true
		}
		kept := map[uint]bool{}
		for _, p := range req.CustomParameter {
			// an id of another device's parameter is treated as new
			if p.ID != 0 && existing[p.ID] {
				kept[p.ID] = true
				if err := tx.Model(&models.CustomParameter{}).Where("id = ? AND device_id = ?", p.ID, device.ID).
					Updates(map[string]interface{}{"key": p.Key, "value": p.Value}).Error; err != nil {
					return err
				}
				continue
			}
			if err := tx.Create(&models.CustomParameter{DeviceID: device.ID, Key: p.Key, Value: p.Value}).Error; err != nil {
				return err
			}
		}
		var removed []uint
		for id := range existing {
			if !kept[id] {
				removed = append(removed, id)
			}
		}
		if len(removed) > 0 {
			if err := tx.Delete(&models.CustomParameter{}, "id IN ? AND device_id = ?", removed, device.ID).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	updated, ok := s.findProjectDevice(c, project, "ConnectionParameter", "DeviceParameter", "CustomParameter")
	if !ok {
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (s *Service) deviceRunTests(c *gin.Context, project *models.Project) {
	type Request struct {
		TestName string `json:"testName"`
		Env      string `json:"env"`
	}
	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	var req Request
	c.Bind(&req)
	arr := strings.Split(req.TestName, " ")
	if len(arr) != 2 {
		s.error(c, http.StatusNotFound, fmt.Errorf("invalid method signature"))
		return
	}

	device, ok := s.findProjectDevice(c, project)
	if !ok {
		return
	}

	dev, _ := s.devicesManager.GetDevice(device.DeviceIdentifier)
	if dev == nil {
		s.error(c, http.StatusNotFound, fmt.Errorf("real device not found"))
		return
	}

	if err := events.LockDevice(dev); err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}
	defer func() {
		_ = events.UnlockDevice(dev)
	}()

	if !dev.IsAppConnected() {
		s.error(c, http.StatusNotFound, fmt.Errorf("no app connected"))
		return
	}
	if false {
		var testsAction action.TestsGet
		s.devicesManager.SendAction(dev, &testsAction)
		logrus.Infof("tests %v", testsAction)
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
