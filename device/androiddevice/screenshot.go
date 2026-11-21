package androiddevice

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png" // registers the PNG decoder for image.DecodeConfig
	"os"

	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

// notReady says why adb can't use the device right now, or nil when it can.
func (d *Device) notReady() error {
	switch state := d.DeviceState(); state {
	case device.StateBooted:
		return nil
	case device.StateRemoteDisconnected:
		return fmt.Errorf("device %s is disconnected: adb on the node doesn't list it (check the USB cable and that the phone is unlocked; adb devices -l on the node should show it as \"device\")", d.DeviceID())
	case device.StateUnauthorized:
		return fmt.Errorf("device %s is unauthorized: accept \"Allow USB debugging?\" on the phone", d.DeviceID())
	default:
		return fmt.Errorf("device %s is %s", d.DeviceID(), device.StateToString(state))
	}
}

// GetScreenshot implements device.Screenshotter.
func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	// don't run adb (and log its errors) for devices adb can't use right now
	if err := d.notReady(); err != nil {
		return nil, 0, 0, err
	}
	d.Log("device", "Get Native Screenshot")
	start := d.deps.Clock.Now()
	data, err := d.screencap()
	if err != nil {
		return nil, 0, 0, err
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("decode screenshot: %w", err)
	}
	logrus.Debugf("Android take screenshot took: %d ms", d.deps.Clock.Now().Sub(start).Milliseconds())
	return data, cfg.Width, cfg.Height, nil
}

// screencap streams the PNG with exec-out and falls back to a file transfer.
func (d *Device) screencap() ([]byte, error) {
	data, err := d.screencapExecOut()
	if err == nil {
		return data, nil
	}
	d.Log("device", "exec-out screencap failed (%v), falling back to file transfer", err)
	return d.screencapViaFile()
}

// screencapExecOut streams the PNG over a single binary-safe adb call.
func (d *Device) screencapExecOut() ([]byte, error) {
	data, err := d.run("exec-out", "screencap", "-p")
	if err != nil {
		return nil, err
	}
	if _, _, err := image.DecodeConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("invalid png from exec-out: %w", err)
	}
	return data, nil
}

// screencapViaFile is the fallback for devices without a binary-safe exec-out.
func (d *Device) screencapViaFile() ([]byte, error) {
	fileName := d.deps.NewID() + ".png"
	remote := "/sdcard/" + fileName
	if _, err := d.run("shell", "screencap", "-p", remote); err != nil {
		return nil, err
	}
	defer func() { _, _ = d.run("shell", "rm", remote) }()
	if _, err := d.run("pull", remote, fileName); err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(fileName) }()
	return os.ReadFile(fileName)
}

// getScreenXml dumps the view hierarchy with uiautomator.
func (d *Device) getScreenXml() (*xmlquery.Node, error) {
	start := d.deps.Clock.Now()
	defer func() {
		logrus.Infof("Android getScreenXml took: %d ms", d.deps.Clock.Now().Sub(start).Milliseconds())
	}()
	out, err := d.run("exec-out", "uiautomator", "dump", "/dev/tty")
	if err != nil {
		return nil, err
	}
	return xmlquery.Parse(bytes.NewReader(out))
}
