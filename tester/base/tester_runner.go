package base

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/protocol"
	"gorm.io/gorm"
	"net"
)

type DeviceMap struct {
	Device device.Device
	Model  models.Device
}

type TestRunner struct {
	DeviceManager manager.Devices
	IP            net.IP
	DB            *gorm.DB
	Config        models.TestConfig
	Test          models.Test

	ProtocolWriter *protocol.ProtocolWriter
	TestRun        models.TestRun
	Err            error

	ProjectId string
	AppId     uint
}

func (t *TestRunner) Init(deviceManager manager.Devices, ip net.IP, db *gorm.DB, projectId string, appId uint) {
	t.DeviceManager = deviceManager
	t.IP = ip
	t.DB = db
	t.ProjectId = projectId
	t.AppId = appId
}
