package androiddevice

import (
	"fmt"
	"github.com/fsuhrau/automationhub/tools/exec"
)

func (d *Device) pressKey(key int) error {
	cmd := exec.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "keyevent", fmt.Sprintf("%d", key))
	return cmd.Run()
}

func (d *Device) swipe(fromX, fromY, toX, toY int) error {
	cmd := exec.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "swipe", fmt.Sprintf("%d", fromX), fmt.Sprintf("%d", fromY), fmt.Sprintf("%d", toX), fmt.Sprintf("%d", toY))
	return cmd.Run()
}

func (d *Device) sendText(text string) error {
	cmd := exec.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "text", fmt.Sprintf("'%s", text))
	return cmd.Run()
}
