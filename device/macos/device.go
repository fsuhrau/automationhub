package macos

import (
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/storage/models"
	exec2 "github.com/fsuhrau/automationhub/tools/exec"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/app"

	"github.com/fsuhrau/automationhub/device"
)

var (
	AppInstallPath   string
	ScreenshotTmpDir string
)

func init() {
	AppInstallPath = filepath.Join(os.Getenv("HOME"), ".automationhub", "installs")
	os.MkdirAll(AppInstallPath, os.ModePerm)
	ScreenshotTmpDir = filepath.Join(os.TempDir(), "automation_hub")
	os.MkdirAll(ScreenshotTmpDir, os.ModePerm)
}

const ConnectionTimeout = 10 * time.Second

type Device struct {
	generic.Device
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
	parameter          string
	deviceParameter    map[string]string
}

func (d *Device) DeviceParameter() map[string]string {
	return d.deviceParameter
}

func (d *Device) DeviceModel() string {
	return d.deviceModel
}

func (d *Device) DeviceType() int {
	return int(models.DeviceTypeDesktop)
}
func (d *Device) PlatformType() int {
	return int(models.PlatformTypeMac)
}

func (d *Device) Parameter() string {
	data, _ := json.Marshal(d.deviceParameter)
	return string(data)
}

func UnpackDeviceParameter(params string) []models.DeviceParameter {
	var parameters []models.DeviceParameter
	deviceParams := make(map[string]string)
	if err := json.Unmarshal([]byte(params), &deviceParams); err == nil {
		for k, v := range deviceParams {
			parameters = append(parameters, models.DeviceParameter{
				Key:   k,
				Value: v,
			})
		}
	}
	return parameters
}

func (d *Device) DeviceOSName() string {
	return d.deviceOSName
}

func (d *Device) DeviceOSVersion() string {
	return d.deviceOSVersion
}

func (d *Device) TargetVersion() string {
	return ""
}

func (d *Device) DeviceName() string {
	return d.deviceName
}

func (d *Device) DeviceID() string {
	return d.deviceID
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

	d.deviceParameter = make(map[string]string)
	d.deviceParameter["SerialNumber"] = d.deviceSerialNumber
	if cpuInfo, err := GetCPUInfo(); err == nil {
		d.deviceParameter["CPU"] = cpuInfo
	}

	if gpuInfo, err := GetGPUInfo(); err == nil {
		d.deviceParameter["GPU"] = gpuInfo
	}

	if ramInfo, err := GetRAMInfo(); err == nil {
		d.deviceParameter["RAM"] = ramInfo
	}

	if supporedEngins, err := GetSupportedGraphicsEngines(); err == nil {
		d.deviceParameter["Driver"] = strings.Join(supporedEngins, ",")
	}

	return nil
}

func (d *Device) IsAppInstalled(params *app.Parameter) (bool, error) {
	appDir := filepath.Join(AppInstallPath, params.App.Hash, params.Name)

	info, err := os.Stat(appDir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (d *Device) InstallApp(params *app.Parameter) error {
	appDir := filepath.Join(AppInstallPath, params.App.Hash)
	_, err := app.Unzip(params.App.AppPath, appDir)
	if err != nil {
		return err
	}

	executable := filepath.Join(AppInstallPath, params.App.Hash, params.Name, params.App.Executable.Executable)
	return os.Chmod(executable, 0755)
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	return os.RemoveAll(filepath.Join(AppInstallPath, params.App.Hash))
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, nodeUrl string) error {
	appDir := filepath.Join(AppInstallPath, params.App.Hash, params.Name)
	executable := filepath.Join(appDir, params.App.Executable.Executable)
	d.runningExecutable = filepath.Base(executable)
	cmd := exec2.NewCommand(executable, "--sessionId="+sessionId, "NODE_URL", "--nodeURL="+nodeUrl, "--deviceId="+d.deviceID)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	executable := filepath.Base(params.App.Executable.Executable)

	d.runningExecutable = ""

	cmd := exec2.NewCommand("killall", executable)
	return cmd.Run()
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

// getWindowID retrieves the window ID of the specified process
func getWindowID(processName string) (string, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("osascript -e 'tell application \"%s\" to id of window 1'", processName))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// takeScreenshot captures the screen of the specified window ID
func takeScreenshot(windowID, filePath string) error {
	cmd := exec.Command("screencapture", "-l", windowID, filePath)
	return cmd.Run()
}

func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	id, err := getWindowID(d.runningExecutable)
	if err != nil {
		return nil, 0, 0, err
	}

	tmpFile := filepath.Join(ScreenshotTmpDir, id+".png")

	if err := takeScreenshot(id, tmpFile); err != nil {
		return nil, 0, 0, err
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, 0, 0, err
	}

	return data, 0, 0, nil
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
