package androiddevice

import (
	"errors"
	"regexp"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
)

// unlockScreen wakes the screen and enters the PIN when the device is locked.
func (d *Device) unlockScreen(config *device.DeviceConfig) error {
	d.Log("device", "Unlock Screen")
	isAwake, err := d.IsAwake()
	if err != nil {
		return err
	}
	if !isAwake {
		d.wakeUp()
	}
	isLocked, err := d.IsPinLocked()
	if err != nil {
		return err
	}
	if isLocked {
		d.enterPin(device.GetAttribute(config, generic.AttributePin))
	}
	return nil
}

func (d *Device) wakeUp() {
	_ = d.pressKey(KEYCODE_WAKEUP)
	_ = d.pressKey(KEYCODE_MENU)
	_ = d.Swipe(400, 800, 400, 200, 0)
}

// enterPin types the digits of the PIN and confirms with enter.
func (d *Device) enterPin(pin string) {
	if pin == "" {
		return
	}
	for _, digit := range pin {
		_ = d.pressKey(KEYCODE_NUMPAD_0 + int(digit-'0'))
	}
	_ = d.pressKey(KEYCODE_NUMPAD_ENTER)
}

// IsPinLocked reports whether the lock screen is showing.
func (d *Device) IsPinLocked() (bool, error) {
	return d.dumpsysFlag(DeviceLockRegex, "window", "mDreamingLockscreen")
}

// IsAwake reports whether the display is on.
func (d *Device) IsAwake() (bool, error) {
	return d.dumpsysFlag(DeviceAwakeRegex, "power", "mHoldingDisplaySuspendBlocker")
}

// dumpsysFlag reads a true/false field of a dumpsys service. It reports true
// together with the error when the field can't be read.
func (d *Device) dumpsysFlag(re *regexp.Regexp, service, field string) (bool, error) {
	out, err := d.run("shell", "dumpsys", service, "|", "grep", field)
	if err != nil {
		return true, err
	}
	findings := re.FindStringSubmatch(string(out))
	if len(findings) == 0 {
		return true, errors.New(field + " can't be evaluated")
	}
	return findings[1] == "true", nil
}
