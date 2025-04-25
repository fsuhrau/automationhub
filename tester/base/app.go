package base

import (
	"context"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/utils/sync"
	"time"
)

type DeviceFunc = func(d device.Device)

func (tr *TestRunner) InstallApp(ctx context.Context, params app.Parameter, devices []DeviceMap) {

	wg := sync.NewExtendedWaitGroup(ctx)

	for _, d := range devices {
		wg.Add(1)
		go func(appp app.Parameter, d device.Device, group sync.ExtendedWaitGroup) {
			defer group.Done()

			/*
				available, err := d.IsAppBundleAvailable(&params)
				if err != nil {
					tr.LogError("check installation failed: %v", err)
					return
				}
			*/

			installed, err := d.IsAppInstalled(&params)
			if err != nil {
				tr.LogError("check installation failed: %v", err)
				return
			}

			if !installed {
				tries := 0
				for {
					if tries > 1 {
						tr.LogError("unable to install app: %v", err)
						break
					}
					if err := d.InstallApp(&appp); err != nil {
						tr.LogInfo("installation failed try delete first: %v", err)
						d.UninstallApp(&appp)
						tries++
						continue
					}
					break
				}
			}
		}(params, d.Device, wg)
	}
	wg.Wait()
}

func (tr *TestRunner) StopApp(ctx context.Context, params app.Parameter, devices []DeviceMap) {

	wg := sync.NewExtendedWaitGroup(ctx)

	for _, d := range devices {
		wg.Add(1)
		go func(dm manager.Devices, param app.Parameter, d device.Device, group sync.ExtendedWaitGroup) {
			if err := d.StopApp(&param); err != nil {
				tr.LogError("unable to stop app: %v", err)
			}
			group.Done()
		}(tr.DeviceManager, params, d.Device, wg)
	}
	wg.Wait()
}

func (tr *TestRunner) StartApp(ctx context.Context, params app.Parameter, devices []DeviceMap, appStartedFunc DeviceFunc, connectedFunc DeviceFunc) ([]DeviceMap, error) {

	cancelContext, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.NewExtendedWaitGroup(cancelContext)
	connectedChan := make(chan DeviceMap, len(devices)) // Buffered channel to collect connected devices

	for _, d := range devices {
		wg.Add(1)
		go tr.startApp(cancelContext, wg, connectedChan, params, d, tr.ProtocolWriter.SessionID(), appStartedFunc, connectedFunc)
	}
	err := wg.WaitWithTimeout(1 * time.Minute)

	close(connectedChan)
	var connectedDevices []DeviceMap
	for d := range connectedChan {
		connectedDevices = append(connectedDevices, d)
	}

	return connectedDevices, err
}

func (tr *TestRunner) startApp(context context.Context, group sync.ExtendedWaitGroup, deviceChan chan DeviceMap, appp app.Parameter, d DeviceMap, sessionId string, appStartedFunc DeviceFunc, connectedFunc DeviceFunc) {
	waitTime := 500 * time.Millisecond
	waitTrigger := time.NewTimer(waitTime)
	defer group.Done()
	startTime := time.Now()
	defer waitTrigger.Stop()

	tr.LogInfo("Start App '%s' on Device '%s' with Session '%s'", appp.Identifier, d.Device.DeviceID(), tr.ProtocolWriter.SessionID())
	if err := d.Device.StartApp(nil, &appp, tr.ProtocolWriter.SessionID(), tr.NodeUrl); err != nil {
		tr.LogError("unable to start app: %v", err)
		return
	}

	if appStartedFunc != nil {
		appStartedFunc(d.Device)
	}

	for {
		select {
		case <-context.Done():
			_ = d.Device.StopApp(&appp)
			tr.LogInfo("Context cancelled, stopping app start process")
			return
		case <-waitTrigger.C:
			if d.Device.IsAppConnected() {
				break
			}
			waitTrigger.Reset(waitTime)
			continue
		}
		break
	}
	tr.ProtocolWriter.TrackStartupTime(d.Model.ID, time.Now().Sub(startTime).Milliseconds())
	if connectedFunc != nil {
		connectedFunc(d.Device)
	}

	deviceChan <- d
}
