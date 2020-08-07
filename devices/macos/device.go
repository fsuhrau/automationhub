package macos

import (
	"github.com/fsuhrau/automationhub/app"
	"net"
	"os/exec"
	"time"

	"github.com/fsuhrau/automationhub/devices"
)

var reinstall bool = false

const CONNECTION_TIMEOUT = 10 * time.Second

type Device struct {
	deviceOSName    string
	deviceOSVersion string
	deviceName      string
	deviceID        string
	deviceIP        net.IP
	deviceState     devices.State
	connectionState devices.ConnectionState

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

func (d *Device) DeviceState() devices.State {
	return d.deviceState
}

func (d *Device) SetDeviceState(state string) {
	switch state {
	case "Booted":
		d.deviceState = devices.Booted
	case "Shutdown":
		d.deviceState = devices.Shutdown
	default:
		d.deviceState = devices.Unknown
	}
}

func (d *Device) UpdateParameter() error {
	return nil
}

func (d *Device) IsAppInstalled(params *app.Parameter) bool {
	return true
}

func (d *Device) InstallApp(params *app.Parameter) error {
	return nil
}

func (d *Device) UninstallApp(bundleId string) error {
	return nil
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {
	if reinstall {
		cmd := devices.NewCommand("open", params.AppPath, "SESSION_ID", sessionId, "HOST", hostIP.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	if reinstall {
		cmd := devices.NewCommand("killall", params.AppPath)
		return cmd.Run()
	}
	return nil
}

func (d *Device) IsAppConnected() bool {
	return d.connectionState == devices.Connected
}

func (d *Device) SetConnectionState(state devices.ConnectionState) {
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
