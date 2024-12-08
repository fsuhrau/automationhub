package base

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/protocol"
	"gorm.io/gorm"
)

type DeviceMap struct {
	Device device.Device
	Model  models.Device
}

type TestRunner struct {
	DeviceManager manager.Devices
	NodeUrl       string
	DB            *gorm.DB
	Config        models.TestConfig
	Test          models.Test

	ProtocolWriter *protocol.ProtocolWriter
	TestRun        models.TestRun
	Err            error

	ProjectId string
	AppId     uint
}

func (t *TestRunner) Init(deviceManager manager.Devices, nodeUrl string, db *gorm.DB, projectId string, appId uint) {
	t.DeviceManager = deviceManager
	t.NodeUrl = nodeUrl
	t.DB = db
	t.ProjectId = projectId
	t.AppId = appId
}
