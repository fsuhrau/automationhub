package iossim

import (
	"bytes"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config"
	"image"
	"image/png"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

var restart bool = true

const CONNECTION_TIMEOUT = 10 * time.Second

type Device struct {
	deviceOSName     string
	deviceOSVersion  string
	deviceName       string
	deviceID         string
	deviceIP         net.IP
	deviceState      device.State
	connectionState  device.ConnectionState
	recordingSession *exec.Cmd
	cfg              *config.Device
	lastUpdateAt     time.Time
}

func (d *Device) DeviceOSName() string {
	return d.deviceOSName
}

func (d *Device) DeviceOSVersion() string {
	return d.deviceOSVersion
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
	case "Booted":
		d.deviceState = device.Booted
	case "Shutdown":
		d.deviceState = device.Shutdown
	default:
		d.deviceState = device.Unknown
	}
}

func (d *Device) UpdateDeviceInfos() error {
	return nil
}

func (d *Device) IsAppInstalled(params *app.Parameter) bool {
	cmd := device.NewCommand("xcrun", "simctl", "get_app_container", d.DeviceID(), params.Identifier)
	return cmd.Run() == nil
}

func (d *Device) InstallApp(params *app.Parameter) error {
	cmd := device.NewCommand("xcrun", "simctl", "install", d.DeviceID(), params.AppPath)
	return cmd.Run()
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	cmd := device.NewCommand("xcrun", "simctl", "uninstall", d.DeviceID(), params.Identifier)
	return cmd.Run()
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {
	if restart {
		cmd := device.NewCommand("xcrun", "simctl", "launch", d.DeviceID(), params.Identifier, "SESSION_ID", sessionId, "HOST", hostIP.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	if restart {
		cmd := device.NewCommand("xcrun", "simctl", "terminate", d.DeviceID(), params.Identifier)
		return cmd.Run()
	}
	return nil
}

func (d *Device) IsAppConnected() bool {
	return d.connectionState == device.Connected
}

func (d *Device) SetConnectionState(state device.ConnectionState) {
	d.connectionState = state
}

func (d *Device) StartRecording(path string) error {
	d.recordingSession = device.NewCommand("xcrun", "simctl", "io", d.DeviceID(), "recordVideo", fmt.Sprintf("./%s.mp4", path))
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
	cmd := device.NewCommand("idevicescreenshot", "-u", d.deviceID, fileName)
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
	return CONNECTION_TIMEOUT
}
