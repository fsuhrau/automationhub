package base

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/utils/sync"
	"time"
)

func (tr *TestRunner) InstallApp(params app.Parameter, devices []DeviceMap) {
	var wg sync.ExtendedWaitGroup
	for _, d := range devices {
		installed, err := d.Device.IsAppInstalled(&params)
		if err != nil {
			tr.LogError("check installation failed: %v", err)
			return
		}

		if !installed {
			wg.Add(1)
			go func(appp app.Parameter, d device.Device, group *sync.ExtendedWaitGroup) {
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
				group.Done()
			}(params, d.Device, &wg)
		}
	}
	wg.Wait()
}

func (tr *TestRunner) StopApp(params app.Parameter, devices []DeviceMap) {
	var wg sync.ExtendedWaitGroup
	for _, d := range devices {
		wg.Add(1)
		go func(dm manager.Devices, param app.Parameter, d device.Device, group *sync.ExtendedWaitGroup) {
			if err := d.StopApp(&param); err != nil {
				tr.LogError("unable to stop app: %v", err)
			}
			group.Done()
		}(tr.DeviceManager, params, d.Device, &wg)
	}
	wg.Wait()
}

func (tr *TestRunner) StartApp(params app.Parameter, devices []DeviceMap, appStartedFunc func(d device.Device), connectedFunc func(d device.Device)) error {
	var wg sync.ExtendedWaitGroup
	for _, d := range devices {
		if !d.Device.IsAppConnected() {
			wg.Add(1)
			go func(dm manager.Devices, appp app.Parameter, d device.Device, sessionId string, group *sync.ExtendedWaitGroup) {
				if err := d.StartApp(&appp, tr.ProtocolWriter.SessionID(), tr.IP); err != nil {
					tr.LogError("unable to start app: %v", err)
				}
				if appStartedFunc != nil {
					appStartedFunc(d)
				}
				for !d.IsAppConnected() {
					time.Sleep(500 * time.Millisecond)
				}
				if connectedFunc != nil {
					connectedFunc(d)
				}
				group.Done()
			}(tr.DeviceManager, params, d.Device, tr.ProtocolWriter.SessionID(), &wg)
		}
	}
	return wg.WaitWithTimeout(2 * time.Minute)
}
