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
	"strings"
	"time"
)

func (s *ApiService) getDevices(session *Session, c *gin.Context) {
	var devices []models.Device
	if err := s.db.Find(&devices).Error; err != nil {
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

func (s *ApiService) registerDevices(msg []byte, conn *websocket.Conn, c *gin.Context) {

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

func (s *ApiService) getDeviceStatus(session *Session, c *gin.Context) {

}

func (s *ApiService) deviceRunTests(session *Session, c *gin.Context) {
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
