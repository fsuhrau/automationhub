package androiddevice

import (
	"fmt"

	"github.com/fsuhrau/automationhub/devices"
)

func (d *Device) pressKey(key int) error {
	cmd := devices.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "keyevent", fmt.Sprintf("%d", key))
	return cmd.Run()
}

func (d *Device) swipe(fromX, fromY, toX, toY int) error {
	cmd := devices.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "swipe", fmt.Sprintf("%d", fromX), fmt.Sprintf("%d", fromY), fmt.Sprintf("%d", toX), fmt.Sprintf("%d", toY))
	return cmd.Run()
}
