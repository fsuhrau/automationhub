package base

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/utils/sync"
	"github.com/sirupsen/logrus"
	"slices"
)

// Unavailable says why a device can't run tests right now, or "" when it
// can. A device must be seen by its node, network devices included (they are
// connected with adb beforehand); shut down simulators and editors are
// started by the run.
func Unavailable(d models.Device) string {
	dev, ok := d.Dev.(device.Device)
	if !ok || dev == nil {
		return "not connected to the hub"
	}
	switch dev.DeviceState() {
	case device.StateRemoteDisconnected:
		return "disconnected"
	case device.StateNodeDisconnected:
		return "on a node that is offline"
	case device.StateUnauthorized:
		return "not authorized for USB debugging"
	}
	return ""
}

// deviceLabel names a device in the run log
func deviceLabel(d models.Device) string {
	name := d.Alias
	if name == "" {
		name = d.Name
	}
	if name == "" || name == d.DeviceIdentifier {
		return d.DeviceIdentifier
	}
	return fmt.Sprintf("%s (%s)", name, d.DeviceIdentifier)
}

// LockDevices locks the devices that can run tests; the others are skipped
// with a line in the run log.
func (tr *TestRunner) LockDevices(devs []models.Device) []DeviceMap {
	var devices []DeviceMap
	for _, d := range devs {
		if reason := Unavailable(d); reason != "" {
			tr.LogError("skipping device %s: %s", deviceLabel(d), reason)
			continue
		}
		dev := d.Dev.(device.Device)
		tr.LogInfo("locking device: %s", dev.DeviceID())
		if err := events.LockDevice(dev); err == nil {
			devices = append(devices, DeviceMap{
				Device: dev,
				Model:  d,
			})
			tr.lockedMu.Lock()
			tr.locked = append(tr.locked, d.ID)
			tr.lockedMu.Unlock()
		} else {
			tr.LogError("locking device %s failed: %v", dev.DeviceID(), err)
		}
	}
	return devices
}

func (tr *TestRunner) UnlockDevices(devices []DeviceMap) {
	for i := range devices {
		tr.DeviceManager.Stop(devices[i].Device)
		_ = events.UnlockDevice(devices[i].Device)
		tr.lockedMu.Lock()
		tr.locked = slices.DeleteFunc(tr.locked, func(id uint) bool { return id == devices[i].Model.ID })
		tr.lockedMu.Unlock()
	}
}

// LockedDevices returns the IDs of the devices the run holds right now, so
// the run page can show their screens.
func (tr *TestRunner) LockedDevices() []uint {
	tr.lockedMu.Lock()
	defer tr.lockedMu.Unlock()
	return slices.Clone(tr.locked)
}

func (tr *TestRunner) StartDevices(ctx context.Context, devices []DeviceMap) error {

	wg := sync.NewExtendedWaitGroup(ctx)

	for _, d := range devices {
		switch d.Device.DeviceState() {
		case device.StateUnknown:
			fallthrough
		case device.StateShutdown:
			wg.Add(1)
			go func(dm manager.Devices, d device.Device, group sync.ExtendedWaitGroup) {
				if err := dm.Start(d); err != nil {
					logrus.Errorf("%v", err)
					tr.LogError("unable to start device: %v", err)
				}
				group.Done()
			}(tr.DeviceManager, d.Device, wg)
		case device.StateBooted:
		}
	}
	wg.Wait()
	return nil
}
