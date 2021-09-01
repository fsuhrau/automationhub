package unity

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"gorm.io/gorm"
	"sync"
	"time"
)

type TestRunner struct {
	base.TestRunner
	deviceManager *hub.DeviceManager
	db            *gorm.DB
	test          models.Test
	config        models.TestConfig
	run           models.TestRun
	log           models.TestLog
}

func New(db *gorm.DB) *TestRunner {
	return &TestRunner{
		db: db,
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

func (tr *TestRunner) Run(devs []device.Device, appData models.App) error {
	var devices []device.Device
	for _, d := range devs {
		if err := d.Lock(); err == nil {
			defer func(dev device.Device) {
				_ = dev.Unlock()
			}(d)
			devices = append(devices, d)
		}
	}

	if len(devices) == 0 {
		return fmt.Errorf("no available devices applied")
	}

	sessionID := tr.NewSessionID()
	tr.run = models.TestRun{
		TestID: tr.test.ID,
	}
	if err := tr.db.Create(&tr.run).Error; err != nil {
		return err
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
	tr.log = models.TestLog{
		TestRunID: tr.run.ID,
		StartedAt: time.Now(),
	}
	if err := tr.db.Create(&tr.log).Error; err != nil {
		return err
	}

	var deviceWg sync.WaitGroup
	for _, d := range devices {
		switch d.DeviceState() {
		case device.StateShutdown:
			deviceWg.Add(1)
			go func() {
				tr.deviceManager.Start(d)
				deviceWg.Done()
			}()
		case device.StateBooted:
		}
	}
	deviceWg.Wait()

	appParams := app.Parameter{Identifier: appData.AppID, AppPath: appData.AppPath}
	for _, d := range devices {
		installed, err := d.IsAppInstalled(&appParams)
		if err != nil {
			return err
		}

		if !installed {
			deviceWg.Add(1)
			go func() {
				d.InstallApp(&appParams)
				deviceWg.Done()
			}()
		}
	}
	deviceWg.Wait()

	for _, d := range devices {
		if !d.IsAppConnected() {
			deviceWg.Add(1)
			go func() {
				d.StartApp(&appParams, sessionID, tr.deviceManager.GetHostIP())
				deviceWg.Done()
			}()
		}
	}
	deviceWg.Wait()

	var testList []models.UnityTestFunction
	if tr.config.Unity.RunAllTests {
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
				if err := tr.deviceManager.SendAction(d, a); err != nil {
					return err
				}
			}
		case models.ParallelExecutionType:
			// need to check ranges
			if err := tr.deviceManager.SendAction(devices[deviceIndex], a); err != nil {
				return err
			}
			deviceIndex++
		}
	}

	// TODO wait for results

	return nil
}
