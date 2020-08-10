package macos

import (
	"github.com/fsuhrau/automationhub/app"
	"net"
	"os/exec"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

var reinstall bool = false

const CONNECTION_TIMEOUT = 10 * time.Second

type Device struct {
	deviceOSName    string
	deviceOSVersion string
	deviceName      string
	deviceID        string
	deviceIP        net.IP
	deviceState     device.State
	connectionState device.ConnectionState

	recordingSession *exec.Cmd
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
	return true
}

func (d *Device) InstallApp(params *app.Parameter) error {
	return nil
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	return nil
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {
	if reinstall {
		cmd := device.NewCommand("open", params.AppPath, "SESSION_ID", sessionId, "HOST", hostIP.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	if reinstall {
		cmd := device.NewCommand("killall", params.AppPath)
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
	return nil
}

func (d *Device) StopRecording() error {
	var err error
	return err
}

func (d *Device) ConnectionTimeout() time.Duration {
	return CONNECTION_TIMEOUT
}
