package unity

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/utils/sync"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net"
	"time"
)

type Logger interface {
	Info(string, ...interface{})
	Error(string, ...interface{}) error
	Finish()
}

type ProtocolLogger struct {
	db       *gorm.DB
	protocol models.TestProtocol
}

func NewProtocol(db *gorm.DB, protocol models.TestProtocol) *ProtocolLogger {
	return &ProtocolLogger{db, protocol}
}

func (pl *ProtocolLogger) store(level, message string) {
	logEntry := models.ProtocolEntry{
		TestProtocolID: pl.protocol.ID,
		Timestamp: time.Now(),
		Source: "UnityTestRunner",
		Level: level,
		Message: message,
	}
	pl.db.Create(&logEntry)
	pl.protocol.Entries = append(pl.protocol.Entries, logEntry)
}

func (pl *ProtocolLogger) Info(format string, data ...interface{}) {
	line := fmt.Sprintf(format, data...)
	pl.store("info", line)
}

func (pl *ProtocolLogger) Error(format string, data ...interface{})  error {
	line := fmt.Sprintf(format, data...)
	pl.store("error", line)
	return fmt.Errorf(line)
}

func (pl *ProtocolLogger) Finish() {
	timeNow := time.Now()
	pl.protocol.EndedAt = &timeNow
	pl.db.Updates(&pl.protocol)
}

type TestRunner struct {
	base.TestRunner
	deviceManager manager.Devices
	ip            net.IP
	db            *gorm.DB
	test          models.Test
	config        models.TestConfig
	run           models.TestRun
	log           Logger
}

func New(db *gorm.DB, ip net.IP, deviceManager manager.Devices) *TestRunner {
	return &TestRunner{
		ip:            ip,
		db:            db,
		deviceManager: deviceManager,
	}
}

func (tr *TestRunner) Initialize(test models.Test) error {
	if test.TestConfig.Type != models.TestTypeUnity {
		return fmt.Errorf("config needs to be unity to create a unity test handler")
	}
	tr.test = test
	tr.config = test.TestConfig
	return nil
}

func (tr *TestRunner) newTestSession() error {
	sessionID := tr.NewSessionID()
	tr.run = models.TestRun{
		TestID:    tr.test.ID,
		SessionID: sessionID,
	}
	if err := tr.db.Create(&tr.run).Error; err != nil {
		return err
	}

	log := models.TestProtocol{
		TestRunID: tr.run.ID,
		StartedAt: time.Now(),
	}
	if err := tr.db.Create(&log).Error; err != nil {
		return err
	}

	tr.log = NewProtocol(tr.db, log)

	return nil
}

func (tr *TestRunner) Run(devs []device.Device, appData models.App) error {
	if err := tr.newTestSession(); err != nil {
		return err
	}
	defer tr.testSessionFinished()

	var devices []device.Device
	for _, d := range devs {
		tr.log.Info("locking device: %s", d.DeviceID())
		if err := d.Lock(); err == nil {
			defer func(dev device.Device) {
				_ = dev.Unlock()
			}(d)
			devices = append(devices, d)
		} else {
			tr.log.Error("locking device %s failed: %v", d.DeviceID(), err)
		}
	}

	if len(devices) == 0 {
		return tr.log.Error("no device available")
	}

	/*
		deviceData := models.Device{
			DeviceIdentifier: dev.DeviceID(),
		}
		if err := tr.db.Find(&deviceData).Error; err != nil {
			return err
		}
		parameter := &models.TestParameter{
			TestRunID: tr.run.ID,
			AppID: appData.ID,
			DeviceID: deviceData.ID,
		}
		if err := tr.db.Create(&parameter).Error; err != nil {
			return err
		}
	*/

	tr.log.Info("Starting devices")
	var deviceWg sync.ExtendedWaitGroup
	for _, d := range devices {
		switch d.DeviceState() {
		case device.StateShutdown:
			fallthrough
		case device.StateRemoteDisconnected:
			deviceWg.Add(1)
			go func(dm manager.Devices, d device.Device, group *sync.ExtendedWaitGroup) {
				if err := dm.Start(d); err != nil {
					logrus.Errorf("%v", err)
					tr.log.Error("unable to start device: %v", err)
				}
				group.Done()
			}(tr.deviceManager, d, &deviceWg)
		case device.StateBooted:
		}
	}
	deviceWg.Wait()

	appParams := app.Parameter{
		Identifier:     appData.AppID,
		AppPath:        appData.AppPath,
		LaunchActivity: appData.LaunchActivity,
		Name:           appData.Name,
		Version:        appData.Version,
		Hash:           appData.Hash,
	}

	tr.log.Info("install app on devices")
	for _, d := range devices {
		installed, err := d.IsAppInstalled(&appParams)
		if err != nil {
			return err
		}

		if !installed {
			deviceWg.Add(1)
			go func(appp app.Parameter, d device.Device, group *sync.ExtendedWaitGroup) {
				if err := d.InstallApp(&appp); err != nil {
					tr.log.Error("unable to install app: %v", err)
				}
				group.Done()
			}(appParams, d, &deviceWg)
		}
	}
	deviceWg.Wait()

	tr.log.Info("stop app on devices")
	for _, d := range devices {
		if d.IsAppConnected() {
			deviceWg.Add(1)
			go func(dm manager.Devices, appp app.Parameter, d device.Device, group *sync.ExtendedWaitGroup) {
				if err := d.StopApp(&appp); err != nil {
					tr.log.Error("unable to stop app: %v", err)
				}
				group.Done()
			}(tr.deviceManager, appParams, d, &deviceWg)
		}
	}
	deviceWg.Wait()

	tr.log.Info("start app on devices and wait for connection")
	for _, d := range devices {
		if !d.IsAppConnected() {
			deviceWg.Add(1)
			go func(dm manager.Devices, appp app.Parameter, d device.Device, sessionId string, group *sync.ExtendedWaitGroup) {
				if err := d.StartApp(&appp, tr.run.SessionID, tr.ip); err != nil {
					tr.log.Error("unable to start app: %v", err)
				}
				for !d.IsAppConnected() {
					time.Sleep(500 * time.Millisecond)
				}
				group.Done()
			}(tr.deviceManager, appParams, d, tr.run.SessionID, &deviceWg)
		}
	}
	if err := deviceWg.WaitWithTimeout(30 * time.Second); err == sync.TimeoutError {
		tr.log.Error("one or more apps didn't connect")
		return fmt.Errorf("timout reached")
	}

	var testList []models.UnityTestFunction
	if tr.config.Unity.RunAllTests {
		tr.log.Info("RunAllTests active requesting PlayMode tests")
		a := &action.TestsGet{}
		if err := tr.deviceManager.SendAction(devices[0], a); err != nil {
			return err
		}
		for _, t := range a.Tests {
			testList = append(testList, models.UnityTestFunction{
				Assembly: t.Assembly,
				Class:    t.Class,
				Method:   t.Method,
			})
		}
	}

	// TestClass: "Innium.IntegrationTests.SmokeTests",
	// TestMethod: "System.Collections.IEnumerator ShortSanityTest()",

	deviceIndex := 0
	tr.log.Info("Execute Tests")
	for _, t := range testList {
		a := &action.TestStart{
			Assembly: t.Assembly,
			Class:    t.Class,
			Method:   t.Method,
		}

		// run all tests on all devices
		switch tr.config.ExecutionType {
		case models.SynchronousExecutionType:
			for _, d := range devices {
				tr.log.Info("Run test '%s/%s' on device '%s'", t.Class, t.Method, d.DeviceID())
				if err := tr.deviceManager.SendAction(d, a); err != nil {
					return tr.log.Error("sync send action failed: %v", err)
				}
			}
		case models.ParallelExecutionType:
			// need to check ranges
			tr.log.Info("Run test '%s/%s' on device '%s'", t.Class, t.Method, devices[deviceIndex])
			if err := tr.deviceManager.SendAction(devices[deviceIndex], a); err != nil {
				return tr.log.Error("parallel send action failed: %v", err)
			}
			deviceIndex++
		}
	}

	// TODO wait for results

	return nil
}

func (tr *TestRunner) testSessionFinished() {
	tr.log.Finish()
}
