package unity

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/tester/protocol"
	"github.com/fsuhrau/automationhub/utils/sync"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net"
	"time"
)

type UnityTestRunner struct {
	base.TestRunner
	deviceManager  manager.Devices
	ip             net.IP
	db             *gorm.DB
	test           models.Test
	config         models.TestConfig
	protocolWriter *protocol.ProtocolWriter
	fin            chan bool
}

func New(db *gorm.DB, ip net.IP, deviceManager manager.Devices) *UnityTestRunner {
	return &UnityTestRunner{
		ip:            ip,
		db:            db,
		deviceManager: deviceManager,
		fin:           make(chan bool, 1),
	}
}

func (tr *UnityTestRunner) Initialize(test models.Test) error {
	if test.TestConfig.Type != models.TestTypeUnity {
		return fmt.Errorf("config needs to be unity to create a unity test handler")
	}
	tr.test = test
	tr.config = test.TestConfig
	return nil
}

func (tr *UnityTestRunner) newTestSession() error {
	sessionID := tr.NewSessionID()
	run := &models.TestRun{
		TestID:    tr.test.ID,
		SessionID: sessionID,
	}
	if err := tr.db.Create(run).Error; err != nil {
		return err
	}

	tr.protocolWriter = protocol.NewProtocolWriter(tr.db, run)
	return nil
}

func (tr *UnityTestRunner) logInfo(format string, params ...interface{}) {
	tr.db.Create(&models.TestRunLogEntry{
		TestRunID: tr.protocolWriter.RunID(),
		Level:     "log",
		Log:       fmt.Sprintf(format, params...),
	})
}

func (tr *UnityTestRunner) logError(format string, params ...interface{}) {
	tr.db.Create(&models.TestRunLogEntry{
		TestRunID: tr.protocolWriter.RunID(),
		Level:     "error",
		Log:       fmt.Sprintf(format, params...),
	})
}

func (tr *UnityTestRunner) Run(devs []models.Device, appData models.App) error {
	if err := tr.newTestSession(); err != nil {
		return err
	}
	defer tr.testSessionFinished()

	var devices []device.Device
	for _, d := range devs {
		dev := d.Dev.(device.Device)
		tr.logInfo("locking device: %s", dev.DeviceID())
		if err := dev.Lock(); err == nil {
			dev.SetActionHandler(tr)
			defer func(dev device.Device) {
				dev.SetActionHandler(nil)
				_ = dev.Unlock()
			}(dev)
			devices = append(devices, dev)
			logWriter, err := tr.protocolWriter.NewProtocol(d.ID, appData.ID)
			if err != nil {
				tr.logError("unable to create LogWriter for %s: %v", dev.DeviceID(), err)
			}
			dev.SetLogWriter(logWriter)
			defer func() {
				dev.SetLogWriter(nil)
			}()
		} else {
			tr.logError("locking device %s failed: %v", dev.DeviceID(), err)
		}
	}

	if len(devices) == 0 {
		tr.logError("no lockable devices available")
		return fmt.Errorf("no lockable devices available")
	}

	tr.logInfo("Starting devices")
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
					tr.logError("unable to start device: %v", err)
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

	tr.logInfo("install app on devices")
	for _, d := range devices {
		installed, err := d.IsAppInstalled(&appParams)
		if err != nil {
			return err
		}

		if !installed {
			deviceWg.Add(1)
			go func(appp app.Parameter, d device.Device, group *sync.ExtendedWaitGroup) {
				if err := d.InstallApp(&appp); err != nil {
					tr.logError("unable to install app: %v", err)
				}
				group.Done()
			}(appParams, d, &deviceWg)
		}
	}
	deviceWg.Wait()

	tr.logInfo("stop app on devices")
	for _, d := range devices {
		if d.IsAppConnected() {
			deviceWg.Add(1)
			go func(dm manager.Devices, appp app.Parameter, d device.Device, group *sync.ExtendedWaitGroup) {
				if err := d.StopApp(&appp); err != nil {
					tr.logError("unable to stop app: %v", err)
				}
				group.Done()
			}(tr.deviceManager, appParams, d, &deviceWg)
		}
	}
	deviceWg.Wait()

	tr.logInfo("start app on devices and wait for connection")
	for _, d := range devices {
		if !d.IsAppConnected() {
			deviceWg.Add(1)
			go func(dm manager.Devices, appp app.Parameter, d device.Device, sessionId string, group *sync.ExtendedWaitGroup) {
				if err := d.StartApp(&appp, tr.protocolWriter.SessionID(), tr.ip); err != nil {
					tr.logError("unable to start app: %v", err)
				}
				for !d.IsAppConnected() {
					time.Sleep(500 * time.Millisecond)
				}
				group.Done()
			}(tr.deviceManager, appParams, d, tr.protocolWriter.SessionID(), &deviceWg)
		}
	}
	if err := deviceWg.WaitWithTimeout(30 * time.Second); err == sync.TimeoutError {
		tr.logError("one or more apps didn't connect")
		return fmt.Errorf("timout reached")
	}

	var testList []models.UnityTestFunction
	if tr.config.Unity.RunAllTests {
		tr.logInfo("RunAllTests active requesting PlayMode tests")
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
	} else {
		tr.db.Where("test_config_unity_id = ?", tr.config.Unity.ID).Find(&testList)
	}

	// TestClass: "Innium.IntegrationTests.SmokeTests",
	// TestMethod: "System.Collections.IEnumerator ShortSanityTest()",

	deviceIndex := 0
	tr.logInfo("Execute Tests")
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
				tr.logInfo("Run test '%s/%s' on device '%s'", t.Class, t.Method, d.DeviceID())
				if err := tr.deviceManager.SendAction(d, a); err != nil {
					tr.logError("sync send action failed: %v", err)
					return fmt.Errorf("sync send action failed: %v", err)
				}
			}
		case models.ParallelExecutionType:
			// need to check ranges
			tr.logInfo("Run test '%s/%s' on device '%s'", t.Class, t.Method, devices[deviceIndex])
			if err := tr.deviceManager.SendAction(devices[deviceIndex], a); err != nil {
				tr.logError("parallel send action failed: %v", err)
				return fmt.Errorf("parallel send action failed: %v", err)
			}
			deviceIndex++
		}
	}

	for {
		select {
		case blah := <-tr.fin:
			{
				if blah {
					logrus.Debug("blubb")
					break
				}
			}
		}
	}

	return nil
}

func (tr *UnityTestRunner) OnActionResponse(d interface{}, response *action.Response) {
	dev := d.(device.Device)
	if response == nil {
		tr.fin <- true
		dev.Log("UnityTestRunner Device Disconnected")
		return
	}

	dev.Log("UnityTestRunner Action Response: %v", response)
	if response.ActionType == action.ActionType_Log {
		if response.GetValue() == "End" {
			tr.fin <- true
		}
	}
}

func (tr *UnityTestRunner) testSessionFinished() {
	tr.protocolWriter.Close()
}
