package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Selectable struct {
	Name             string
	DeviceIdentifier string
	OS               string
	OSVersion        string
	Status           string
	Connected        *action.Connect
}
type Selectables []*Selectable

func (s Selectables) Len() int      { return len(s) }
func (s Selectables) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ Selectables }

func (s ByName) Less(i, j int) bool { return s.Selectables[i].Name < s.Selectables[j].Name }

func (s *ApiService) getDevices(session *Session, c *gin.Context) {
	var devices []models.Device
	if err := s.db.Find(&devices).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	for i := range devices {
		devices[i].Dev = s.devicesManager.GetDevice(devices[i].DeviceIdentifier)
	}

	c.JSON(http.StatusOK, devices)
}

func (s *ApiService) getDeviceStatus(session *Session, c *gin.Context) {

}

func (s *ApiService) runTests(session *Session, c *gin.Context) {
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
	} else {

		runTestAction := action.TestStart{
			Class: "Innium.IntegrationTests.SmokeTests",
		}
		s.devicesManager.SendAction(dev, &runTestAction)
	}
}
