package web

import (
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/storage/models"
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
	browser         string
	browserPath     string
	browserVersion  string
	deviceOSName    string
	deviceOSVersion string
	deviceName      string
	deviceID        string
	deviceIP        net.IP
	deviceState     device.State
	lastUpdateAt    time.Time
	deviceParameter map[string]string

	runningExecutable  string
	recordingSession   *exec.Cmd
	applicationProcess *exec.Cmd
}

func (d *Device) DeviceParameter() map[string]string {
	return d.deviceParameter
}

func (d *Device) DeviceType() int {
	return int(models.DeviceTypeBrowser)
}

func (d *Device) PlatformType() int {
	return int(models.PlatformTypeWeb)
}

func (d *Device) DeviceOSName() string {
	return d.deviceOSName
}

func (d *Device) DeviceOSVersion() string {
	return d.deviceOSVersion
}

func (d *Device) DeviceName() string {
	if len(d.browserVersion) > 1 {
		return fmt.Sprintf("%s (%s) @ %s", d.browser, d.browserVersion, d.deviceName)
	}
	return fmt.Sprintf("%s @ %s", d.browser, d.deviceName)
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

func (d *Device) UpdateDeviceInfos() {
	d.deviceOSName, _ = GetOSName()
	d.deviceName, _ = GetDeviceName()
	d.deviceOSVersion, _ = GetOSVersion()
	d.deviceID, _ = GetHardwareUUID()
	d.deviceParameter = make(map[string]string)
	if modelNumber, err := GetModelNumber(); err == nil {
		d.deviceParameter["Device Model"] = modelNumber
	}

	if serialNumber, err := GetSerialNumber(); err == nil {
		d.deviceParameter["Serial Number"] = serialNumber
	}

	var versionCommand *exec.Cmd
	if strings.Contains(runtime.GOOS, "windows") {
		versionCommand = exec2.NewCommand(d.browserPath, "--version")
	} else if strings.Contains(runtime.GOOS, "darwin") {
		if d.browser == "safari" {
			versionCommand = exec.Command("defaults", "read", d.browserPath+"/Contents/Info", "CFBundleShortVersionString")
		} else {
			versionCommand = exec2.NewCommand(d.browserPath, "--version")
		}
	} else if strings.Contains(runtime.GOOS, "linux") {
		versionCommand = exec2.NewCommand(d.browserPath, "--version")
	}
	data, err := versionCommand.Output()
	if err != nil {
		fmt.Println("Failed to get version for ", d.browser)
	} else {
		versionString := string(data)
		versionString = strings.TrimSpace(versionString)
		index := strings.LastIndex(versionString, " ")
		if index != -1 {
			d.browserVersion = versionString[index+1:]
		} else {
			d.browserVersion = versionString
		}
		d.deviceParameter["Browser Version"] = d.browserVersion
	}
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
