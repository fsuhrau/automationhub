package hub

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/node"
)

func (dm *DeviceManager) Run(ctx context.Context, runSocketListener bool) error {
	dm.log.Debug("Starting device manager")

	dm.notifier = dm.newNotifier(dm.deviceSnapshot)
	dm.startHandlers()
	// the initial list, so the master is up to date without asking
	dm.notifier.Notify()

	go dm.refreshLoop(ctx)
	return nil
}

func (dm *DeviceManager) startHandlers() {
	for _, v := range dm.deviceHandlers {
		if n, ok := v.(device.ChangeNotifying); ok {
			n.SetChangeNotifier(dm.notifier.Notify)
		}
		if err := v.Init(dm.masterUrl, dm.nodeIdentifier, dm.authToken); err != nil {
			dm.log.Errorf("init handler %s: %v", v.Name(), err)
		}
		if err := v.Start(); err != nil {
			dm.log.Errorf("start handler %s: %v", v.Name(), err)
		}
	}
}

func (dm *DeviceManager) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(dm.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			dm.log.Infof("Stop DeviceManager")
			return
		case <-ticker.C:
			dm.refreshDeviceList()
		case <-dm.refreshSignal:
			dm.refreshDeviceList()
		}
	}
}

func (dm *DeviceManager) refreshDeviceList() {
	for _, m := range dm.deviceHandlers {
		if err := m.RefreshDevices(false); err != nil {
			dm.log.Debugf("refresh devices failed for manager %s: %v", m.Name(), err)
		}
	}
	// handlers without OS events (simulators, editors, ...) and state changes
	// made by sessions only show up here; the notifier sends nothing when the
	// list is unchanged
	dm.notifier.Notify()
}

func (dm *DeviceManager) deviceSnapshot() []*node.DeviceResponse {
	devices, _ := dm.Devices()
	return node.ToDevicesResponse(devices)
}

func (dm *DeviceManager) TriggerRefresh() {
	select {
	case dm.refreshSignal <- struct{}{}:
	default: // Avoid blocking if the channel is full
	}
}
