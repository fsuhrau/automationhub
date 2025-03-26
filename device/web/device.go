package web

import (
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	exec2 "github.com/fsuhrau/automationhub/tools/exec"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/app"

	"github.com/fsuhrau/automationhub/device"
)

var (
	ScreenshotTmpDir string
)

func init() {
	ScreenshotTmpDir = filepath.Join(os.TempDir(), "automation_hub")
	os.MkdirAll(ScreenshotTmpDir, os.ModePerm)
}

const ConnectionTimeout = 10 * time.Second

type Device struct {
	generic.Device
	browser            string
	browserPath        string
	browserVersion     string
	deviceOSName       string
	deviceOSVersion    string
	deviceName         string
	deviceID           string
	deviceIP           net.IP
	deviceState        device.State
	recordingSession   *exec.Cmd
	lastUpdateAt       time.Time
	deviceModel        string
	deviceSerialNumber string
	runningExecutable  string

	applicationProcess *exec.Cmd
}

func (d *Device) DeviceModel() string {
	return d.deviceModel
}

func (d *Device) DeviceOSName() string {
	return d.deviceOSName
}

func (d *Device) DeviceOSVersion() string {
	return d.deviceOSVersion
}

func (d *Device) DeviceName() string {
	return d.browser + " @ " + d.deviceName
}

func (d *Device) DeviceID() string {
	return d.deviceID + "/" + d.browser
}

func (d *Device) TargetVersion() string {
	return d.browserVersion
}

func (d *Device) DeviceIP() net.IP {
	return d.deviceIP
}

func (d *Device) DeviceState() device.State {
	return d.deviceState
}

func (d *Device) SetDeviceState(state string) {
	switch state {
	case "StateBooted":
		d.deviceState = device.StateBooted
	case "StateShutdown":
		d.deviceState = device.StateShutdown
	default:
		d.deviceState = device.StateUnknown
	}
}

func (d *Device) UpdateDeviceInfos() error {
	var err error

	if d.deviceOSName, err = GetOSName(); err != nil {
		return err
	}

	if d.deviceName, err = GetDeviceName(); err != nil {
		return err
	}

	if d.deviceOSVersion, err = GetOSVersion(); err != nil {
		return err
	}

	if d.deviceID, err = GetHardwareUUID(); err != nil {
		return err
	}

	if d.deviceModel, err = GetModelNumber(); err != nil {
		return err
	}

	if d.deviceSerialNumber, err = GetSerialNumber(); err != nil {
		return err
	}

	return nil
}

func (d *Device) IsAppInstalled(params *app.Parameter) (bool, error) {
	return true, nil
}

func (d *Device) InstallApp(params *app.Parameter) error {
	return nil
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	return nil
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, nodeUrl string) error {
	applicationURL := fmt.Sprintf("%s?sessionId=%s&nodeURL=%s&deviceId=%s", params.Web.StartURL, sessionId, nodeUrl, d.DeviceID())
	if strings.Contains(runtime.GOOS, "windows") {
		d.applicationProcess = exec2.NewCommand(d.browserPath, applicationURL)
	} else if strings.Contains(runtime.GOOS, "darwin") {
		d.applicationProcess = exec2.NewCommand(d.browserPath, applicationURL)
	} else if strings.Contains(runtime.GOOS, "linux") {
		d.applicationProcess = exec2.NewCommand(d.browserPath, applicationURL)
	}

	if d.applicationProcess == nil {
		return fmt.Errorf("no start application process found")
	}

	if err := d.applicationProcess.Start(); err != nil {
		return err
	}

	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	if d.applicationProcess != nil && d.applicationProcess.Process != nil {
		d.applicationProcess.Process.Kill()
	}
	d.applicationProcess = nil
	return nil
}

func (d *Device) IsAppConnected() bool {
	return d.Connection() != nil
}

func (d *Device) StartRecording(path string) error {
	return nil
}

func (d *Device) StopRecording() error {
	var err error
	return err
}

func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	return nil, 0, 0, fmt.Errorf("GetScreenshot not Supported by this device")
}

func (d *Device) HasFeature(string) bool {
	return false
}

func (d *Device) Execute(string) {

}

func (d *Device) ConnectionTimeout() time.Duration {
	return ConnectionTimeout
}

func (d *Device) RunNativeScript(script []byte) {

}
