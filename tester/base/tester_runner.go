package base

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/protocol"
	"gorm.io/gorm"
	"sync"
	"time"
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

	lockedMu sync.Mutex
	locked   []uint

	// closed once the run stopped its apps and released its devices
	finished     chan struct{}
	finishedOnce sync.Once
}

func (t *TestRunner) Init(deviceManager manager.Devices, nodeUrl string, db *gorm.DB, projectId string, appId uint) {
	t.DeviceManager = deviceManager
	t.NodeUrl = nodeUrl
	t.DB = db
	t.ProjectId = projectId
	t.AppId = appId
	t.finished = make(chan struct{})
}

// Finished marks the run as done with its devices; call it when the run's
// work ends, however it ends.
func (t *TestRunner) Finished() {
	t.finishedOnce.Do(func() { close(t.finished) })
}

// WaitFinished waits until the run released its devices, at most timeout.
// A cancel waits for it, so the test can be started again right away: the
// device used to be still locked and its app still connected.
func (t *TestRunner) WaitFinished(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-t.finished:
		return true
	case <-timer.C:
		return false
	}
}
