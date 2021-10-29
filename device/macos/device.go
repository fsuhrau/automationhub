package macos

import (
	"github.com/fsuhrau/automationhub/device/generic"
	exec2 "github.com/fsuhrau/automationhub/tools/exec"
	"net"
	"os/exec"
	"time"

	"github.com/fsuhrau/automationhub/app"

	"github.com/fsuhrau/automationhub/device"
)

const CONNECTION_TIMEOUT = 10 * time.Second

type Device struct {
	generic.Device
	deviceOSName     string
	deviceOSVersion  string
	deviceName       string
	deviceID         string
	deviceIP         net.IP
	deviceState      device.State
	recordingSession *exec.Cmd
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
	case "StateBooted":
		d.deviceState = device.StateBooted
	case "StateShutdown":
		d.deviceState = device.StateShutdown
	default:
		d.deviceState = device.StateUnknown
	}
}

func (d *Device) UpdateDeviceInfos() error {
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

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {
	cmd := exec2.NewCommand("open", params.AppPath, "SESSION_ID", sessionId, "HOST", hostIP.String())
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	cmd := exec2.NewCommand("killall", params.AppPath)
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

func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	return nil, 0, 0, nil
}

func (d *Device) HasFeature(string) bool {
	return false
}

func (d *Device) Execute(string) {

}

func (d *Device) ConnectionTimeout() time.Duration {
	return CONNECTION_TIMEOUT
}

func (d *Device) RunNativeScript(script []byte)  {

}