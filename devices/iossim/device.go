package iossim

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/fsuhrau/automationhub/devices"
)

var restart bool = true

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
	cmd := devices.NewCommand("xcrun", "simctl", "get_app_container", d.DeviceID(), params.Identifier)
	return cmd.Run() == nil
}

func (d *Device) InstallApp(params *app.Parameter) error {
	cmd := devices.NewCommand("xcrun", "simctl", "install", d.DeviceID(), params.AppPath)
	return cmd.Run()
}

func (d *Device) UninstallApp(bundleId string) error {
	cmd := devices.NewCommand("xcrun", "simctl", "uninstall", d.DeviceID(), bundleId)
	return cmd.Run()
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {
	if restart {
		cmd := devices.NewCommand("xcrun", "simctl", "launch", d.DeviceID(), params.Identifier, "SESSION_ID", sessionId, "HOST", hostIP.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	if restart {
		cmd := devices.NewCommand("xcrun", "simctl", "terminate", d.DeviceID(), params.Identifier)
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
	d.recordingSession = devices.NewCommand("xcrun", "simctl", "io", d.DeviceID(), "recordVideo", fmt.Sprintf("./%s.mp4", path))
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

func (d *Device) ConnectionTimeout() time.Duration {
	return CONNECTION_TIMEOUT
}
