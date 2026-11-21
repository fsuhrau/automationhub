package base

import (
	"context"
	"fmt"
	"slices"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/utils/sync"
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
func (r *Runner) LockDevices(devs []models.Device) []DeviceMap {
	var devices []DeviceMap
	for _, d := range devs {
		if dm, ok := r.lockDevice(d); ok {
			devices = append(devices, dm)
		}
	}
	return devices
}

func (r *Runner) lockDevice(d models.Device) (DeviceMap, bool) {
	if reason := Unavailable(d); reason != "" {
		r.LogError("skipping device %s: %s", deviceLabel(d), reason)
		return DeviceMap{}, false
	}
	dev := d.Dev.(device.Device)
	r.LogInfo("locking device: %s", dev.DeviceID())
	if err := events.SetDeviceLock(events.DeviceLockChanged, dev, true); err != nil {
		r.LogError("locking device %s failed: %v", dev.DeviceID(), err)
		return DeviceMap{}, false
	}
	r.lockedMu.Lock()
	r.locked = append(r.locked, d.ID)
	r.lockedMu.Unlock()
	return DeviceMap{Device: dev, Model: d}, true
}

// UnlockDevices stops and unlocks the devices.
func (r *Runner) UnlockDevices(devices []DeviceMap) {
	for _, d := range devices {
		_ = r.Devices.Stop(d.Device)
		_ = events.SetDeviceLock(events.DeviceLockChanged, d.Device, false)
		r.lockedMu.Lock()
		r.locked = slices.DeleteFunc(r.locked, func(id uint) bool { return id == d.Model.ID })
		r.lockedMu.Unlock()
	}
}

// LockedDevices returns the IDs of the devices the run holds right now, so
// the run page can show their screens.
func (r *Runner) LockedDevices() []uint {
	r.lockedMu.Lock()
	defer r.lockedMu.Unlock()
	return slices.Clone(r.locked)
}

// StartDevices starts the devices that are not running, e.g. simulators.
func (r *Runner) StartDevices(ctx context.Context, devices []DeviceMap) {
	forEachDevice(ctx, devices, func(d DeviceMap) {
		switch d.Device.DeviceState() {
		case device.StateUnknown, device.StateShutdown:
			if err := r.Devices.Start(d.Device); err != nil {
				r.LogError("unable to start device: %v", err)
			}
		}
	})
}

// forEachDevice runs fn for every device in parallel and waits until all are
// done or the context ends.
func forEachDevice(ctx context.Context, devices []DeviceMap, fn func(DeviceMap)) {
	wg := sync.NewExtendedWaitGroup(ctx)
	for _, d := range devices {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn(d)
		}()
	}
	_ = wg.Wait()
}
