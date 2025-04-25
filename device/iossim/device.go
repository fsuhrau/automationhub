package iossim

import (
	"bytes"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/storage/models"
	exec2 "github.com/fsuhrau/automationhub/tools/exec"
	"image"
	"image/png"
	"os"
	"os/exec"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

var restart bool = true

const ConnectionTimeout = 10 * time.Second

type Device struct {
	generic.Device
	deviceOSName     string
	deviceOSVersion  string
	deviceName       string
	deviceID         string
	deviceState      device.State
	recordingSession *exec.Cmd
	lastUpdateAt     time.Time
}

func (d *Device) DeviceParameter() map[string]string {
	return nil
}

func (d *Device) DeviceType() int {
	return int(models.DeviceTypePhone)
}
func (d *Device) PlatformType() int {
	return int(models.PlatformTypeiOSSimulator)
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

func (d *Device) IsAppInstalled(params *app.Parameter) (bool, error) {
	cmd := exec2.NewCommand("xcrun", "simctl", "get_app_container", d.DeviceID(), params.Identifier)
	err := cmd.Run()
	return err == nil, err
}

func (d *Device) InstallApp(params *app.Parameter) error {
	cmd := exec2.NewCommand("xcrun", "simctl", "install", d.DeviceID(), params.App.AppPath)
	return cmd.Run()
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	cmd := exec2.NewCommand("xcrun", "simctl", "uninstall", d.DeviceID(), params.Identifier)
	return cmd.Run()
}

func (d *Device) StartApp(_ *device.DeviceConfig, appParams *app.Parameter, sessionId string, nodeUrl string) error {
	if restart {
		cmd := exec2.NewCommand("xcrun", "simctl", "launch", d.DeviceID(), appParams.Identifier, "SESSION_ID", sessionId, "NODE_URL", nodeUrl)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	if restart {
		cmd := exec2.NewCommand("xcrun", "simctl", "terminate", d.DeviceID(), params.Identifier)
		return cmd.Run()
	}
	return nil
}

func (d *Device) IsAppConnected() bool {
	return d.Connection() == nil
}

func (d *Device) StartRecording(path string) error {
	d.recordingSession = exec2.NewCommand("xcrun", "simctl", "io", d.DeviceID(), "recordVideo", fmt.Sprintf("./%s.mp4", path))
	if err := d.recordingSession.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Device) StopRecording() error {
	var err error
	if d.recordingSession != nil && d.recordingSession.Process != nil {
		err = d.recordingSession.Process.Signal(os.Interrupt)
		d.recordingSession = nil
	}
	return err
}

func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	var width int
	var height int
	fileName := fmt.Sprintf("%s.png", d.deviceID)
	cmd := exec2.NewCommand("idevicescreenshot", "-u", d.deviceID, fileName)
	if err := cmd.Run(); err != nil {
		return nil, 0, 0, err
	}

	imagePath, _ := os.Open(fileName)
	defer imagePath.Close()
	srcImage, _, _ := image.Decode(imagePath)

	height = srcImage.Bounds().Dy()
	width = srcImage.Bounds().Dx()

	var data []byte
	writer := bytes.NewBuffer(data)
	err := png.Encode(writer, srcImage)
	return writer.Bytes(), width, height, err
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
