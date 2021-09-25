package unity

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/tester/protocol"
	"github.com/fsuhrau/automationhub/utils/sync"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net"
	"strings"
	"time"
)

type workerChannel chan action.TestStart
type cancelChannel chan bool

type DeviceMap struct {
	Device device.Device
	Model  models.Device
}

type testsRunner struct {
	base.TestRunner
	deviceManager  manager.Devices
	ip             net.IP
	db             *gorm.DB
	test           models.Test
	config         models.TestConfig
	protocolWriter *protocol.ProtocolWriter
	fin            chan bool
	err            error
	publisher      sse.Publisher

	appParams app.Parameter
}

func New(db *gorm.DB, ip net.IP, deviceManager manager.Devices, publisher sse.Publisher) *testsRunner {
	return &testsRunner{
		ip:            ip,
		db:            db,
		deviceManager: deviceManager,
		fin:           make(chan bool, 1),
		publisher:     publisher,
	}
}

func (tr *testsRunner) Initialize(test models.Test) error {
	if test.TestConfig.Type != models.TestTypeUnity {
		return fmt.Errorf("config needs to be unity to create a unity test handler")
	}
	tr.test = test
	tr.config = test.TestConfig
	return nil
}

func (tr *testsRunner) newTestSession(appId uint) error {
	sessionID := tr.NewSessionID()
	run := &models.TestRun{
		TestID:    tr.test.ID,
		AppID:     appId,
		SessionID: sessionID,
	}
	if err := tr.db.Create(run).Error; err != nil {
		return err
	}

	tr.protocolWriter = protocol.NewProtocolWriter(tr.db, run)
	return nil
}

func (tr *testsRunner) logInfo(format string, params ...interface{}) {
	logEntry := &models.TestRunLogEntry{
		TestRunID: tr.protocolWriter.RunID(),
		Level:     "log",
		Log:       fmt.Sprintf(format, params...),
	}
	tr.db.Create(logEntry)
	events.NewTestLogEntry.Trigger(events.NewTestLogEntryPayload{
		logEntry.TestRunID,
		logEntry,
	})
}

func (tr *testsRunner) logError(format string, params ...interface{}) {
	tr.err = fmt.Errorf(format, params)
	logEntry := &models.TestRunLogEntry{
		TestRunID: tr.protocolWriter.RunID(),
		Level:     "error",
		Log:       fmt.Sprintf(format, params...),
	}
	tr.db.Create(logEntry)

	events.NewTestLogEntry.Trigger(events.NewTestLogEntryPayload{
		logEntry.TestRunID,
		logEntry,
	})
}

func (tr *testsRunner) Run(devs []models.Device, appData models.App) error {
	if err := tr.newTestSession(appData.ID); err != nil {
		return err
	}
	defer tr.testSessionFinished()

	var devices []DeviceMap
	for _, d := range devs {
		dev := d.Dev.(device.Device)
		tr.logInfo("locking device: %s", dev.DeviceID())
		if err := dev.Lock(); err == nil {
			defer func(dev device.Device) {
				_ = dev.Unlock()
			}(dev)

			devices = append(devices, DeviceMap{
				Device: dev,
				Model:  d,
			})
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
		switch d.Device.DeviceState() {
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
			}(tr.deviceManager, d.Device, &deviceWg)
		case device.StateBooted:
		}
	}
	deviceWg.Wait()

	tr.appParams = app.Parameter{
		AppID:          appData.ID,
		Identifier:     appData.AppID,
		AppPath:        appData.AppPath,
		LaunchActivity: appData.LaunchActivity,
		Name:           appData.Name,
		Version:        appData.Version,
		Hash:           appData.Hash,
	}

	// stop app
	tr.logInfo("stop apps if running")
	for _, d := range devices {
		deviceWg.Add(1)
		go func(dm manager.Devices, appp app.Parameter, d device.Device, group *sync.ExtendedWaitGroup) {
			if err := d.StopApp(&appp); err != nil {
				tr.logError("unable to stop app: %v", err)
			}
			group.Done()
		}(tr.deviceManager, tr.appParams, d.Device, &deviceWg)
	}
	deviceWg.Wait()

	tr.logInfo("install app on devices")
	for _, d := range devices {
		installed, err := d.Device.IsAppInstalled(&tr.appParams)
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
			}(tr.appParams, d.Device, &deviceWg)
		}
	}
	deviceWg.Wait()

	tr.logInfo("start app on devices and wait for connection")
	for _, d := range devices {
		if !d.Device.IsAppConnected() {
			deviceWg.Add(1)
			go func(dm manager.Devices, appp app.Parameter, d device.Device, sessionId string, group *sync.ExtendedWaitGroup) {
				if err := d.StartApp(&appp, tr.protocolWriter.SessionID(), tr.ip); err != nil {
					tr.logError("unable to start app: %v", err)
				}
				for !d.IsAppConnected() {
					time.Sleep(500 * time.Millisecond)
				}
				group.Done()
			}(tr.deviceManager, tr.appParams, d.Device, tr.protocolWriter.SessionID(), &deviceWg)
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
		if err := tr.deviceManager.SendAction(devices[0].Device, a); err != nil {
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

	tr.logInfo("Execute Tests")

	switch tr.config.ExecutionType {
	case models.SynchronousExecutionType:
		cancel := make(cancelChannel, len(devices))
		group := sync.ExtendedWaitGroup{}

		var workers []workerChannel
		for _, d := range devices {
			channel := make(workerChannel, len(testList))
			workers = append(workers, channel)
			go tr.WorkerFunction(channel, d, cancel, &group)
		}

		for _, t := range testList {
			a := action.TestStart{
				Assembly: t.Assembly,
				Class:    t.Class,
				Method:   t.Method,
			}
			for i := range workers {
				group.Add(1)
				workers[i] <- a
			}
		}
		group.Wait()
		for i := 0; i < len(devices); i++ {
			cancel <- true
		}
		close(cancel)
		for i := range workers {
			close(workers[i])
		}

	case models.ParallelExecutionType:
		cancel := make(cancelChannel, len(devices))
		group := sync.ExtendedWaitGroup{}

		parallelWorker := make(workerChannel, len(testList))
		for _, d := range devices {
			go tr.WorkerFunction(parallelWorker, d, cancel, &group)
		}
		for _, t := range testList {
			a := action.TestStart{
				Assembly: t.Assembly,
				Class:    t.Class,
				Method:   t.Method,
			}
			group.Add(1)
			parallelWorker <- a
		}
		group.Wait()
		for i := 0; i < len(devices); i++ {
			cancel <- true
		}
		close(cancel)
		close(parallelWorker)
	}

	return nil
}

func (tr *testsRunner) WorkerFunction(channel workerChannel, dev DeviceMap, cancel cancelChannel, group *sync.ExtendedWaitGroup) {
	for {
		select {
		case task := <-channel:
			method := task.Method
			methodParts := strings.Split(method, " ")
			if len(methodParts) > 1 {
				method = methodParts[1]
			}
			tr.runTest(dev, task, method)
			group.Done()
		case <-cancel:
			return
		}
	}
}

func (tr *testsRunner) runTest(dev DeviceMap, task action.TestStart, method string) {
	logWriter, err := tr.protocolWriter.NewProtocol(dev.Model.ID, fmt.Sprintf("%s/%s", task.Class, method))
	if err != nil {
		tr.logError("unable to create LogWriter for %s: %v", dev.Device.DeviceID(), err)
	}
	dev.Device.SetLogWriter(logWriter)
	defer func() {
		dev.Device.SetLogWriter(nil)
	}()

	tr.logInfo("Run test '%s/%s' on device '%s'", task.Class, method, dev.Device.DeviceID())
	executor := NewExecuter(tr.deviceManager)
	if err := executor.Execute(dev.Device, task, 2*time.Minute); err != nil {
		tr.logError("test execution failed: %v", err)
	} else {
		tr.logInfo("test execution finished")
	}
}

func (tr *testsRunner) testSessionFinished() {
	tr.protocolWriter.Close()
}
