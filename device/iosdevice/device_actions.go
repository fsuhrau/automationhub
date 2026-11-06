package iosdevice

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

var wdaButtons = map[string]string{
	"home":        "home",
	"volume_up":   "volumeUp",
	"volume_down": "volumeDown",
}

func (d *Device) requireWebDriver() error {
	if d.webDriver == nil {
		return fmt.Errorf("no WebDriverAgent session for device %s", d.DeviceID())
	}
	return nil
}

// pointScale converts screenshot pixels to WDA points.
func (d *Device) pointScale() float64 {
	if d.screenScale > 0 {
		return d.screenScale
	}
	scale, err := d.webDriver.ScreenScale()
	if err != nil || scale <= 0 {
		return 1
	}
	d.screenScale = scale
	return scale
}

// Tap implements device.InputCapable. x and y are screenshot pixels.
func (d *Device) Tap(x, y int) error {
	if err := d.requireWebDriver(); err != nil {
		return err
	}
	scale := d.pointScale()
	return d.webDriver.Tap(float64(x)/scale, float64(y)/scale)
}

// Swipe implements device.InputCapable. Coordinates are screenshot pixels.
func (d *Device) Swipe(fromX, fromY, toX, toY int, duration time.Duration) error {
	if err := d.requireWebDriver(); err != nil {
		return err
	}
	if duration <= 0 {
		duration = 300 * time.Millisecond
	}
	scale := d.pointScale()
	return d.webDriver.Drag(float64(fromX)/scale, float64(fromY)/scale, float64(toX)/scale, float64(toY)/scale, duration.Seconds())
}

// InputText implements device.InputCapable; it types into the focused element.
func (d *Device) InputText(text string) error {
	if err := d.requireWebDriver(); err != nil {
		return err
	}
	return d.webDriver.SendText(text)
}

// PressKey implements device.InputCapable. iOS supports home and volume keys.
func (d *Device) PressKey(key string) error {
	if err := d.requireWebDriver(); err != nil {
		return err
	}
	button, ok := wdaButtons[strings.ToLower(key)]
	if !ok {
		return fmt.Errorf("key %q: %w", key, device.ErrUnsupported)
	}
	return d.webDriver.PressButton(button)
}
