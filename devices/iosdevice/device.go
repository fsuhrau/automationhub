package iosdevice

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/devices"
)

var reinstall bool = true

const CONNECTION_TIMEOUT = 120 * time.Second

type Device struct {
	deviceOSName    string
	deviceOSVersion string
	deviceName      string
	deviceID        string
	deviceState     devices.State
	connectionState devices.ConnectionState
	deviceIP        net.IP

	recordingSessionProcess *exec.Cmd
	runningAppProcess       *exec.Cmd
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

func (d *Device) ExtractAppParameters(bundlePath string) error {
	return nil
}

func (d *Device) IsAppInstalled(bundleId string) bool {
	cmd := devices.NewCommand("/usr/local/bin/ios-deploy", "--id", d.DeviceID(), "--exists", "--bundle_id", bundleId)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "true")
}

func (d *Device) InstallApp(bundlePath string) error {
	cmd := devices.NewCommand("/usr/local/bin/ios-deploy", "--id", d.DeviceID(), "--bundle", bundlePath)
	return cmd.Run()
}

func (d *Device) UninstallApp(bundleId string) error {
	cmd := devices.NewCommand("/usr/local/bin/ios-deploy", "--id", d.DeviceID(), "--uninstall_only", "--bundle_id", bundleId)
	return cmd.Run()
}

func (d *Device) StartApp(appPath string, bundleId string, sessionId string, hostIP net.IP) error {
	if reinstall {
		d.runningAppProcess = devices.NewCommand("/usr/local/bin/ios-deploy", "--json", "--id", d.DeviceID(), "--noinstall", "--noninteractive", "--no-wifi", "--bundle", appPath, "--bundle_id", bundleId, "--args", fmt.Sprintf("SESSION_ID %s HOST %s", sessionId, hostIP.String()))
		if false {
			d.runningAppProcess.Stdout = os.Stdout
		}
		d.runningAppProcess.Stderr = os.Stderr
		if err := d.runningAppProcess.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) StopApp(appPath, bundleId string) error {
	var err error
	if reinstall {
		if d.runningAppProcess != nil && d.runningAppProcess.Process != nil {
			err = d.runningAppProcess.Process.Signal(os.Interrupt)
			if err != nil {
				logrus.Errorf("Stop Interrupt error: %v", err)
			}
			//err = d.runningAppProcess.Process.Kill()
			//if err != nil {
			//	logrus.Errorf("Stop Kill error: %v", err)
			//}
			if e := d.runningAppProcess.Wait(); e != nil{
				logrus.Errorf("Stop Kill error: %v", e)
			}
			d.runningAppProcess = nil
		}
	}
	return err
}

func (d *Device) IsAppConnected() bool {
	return d.connectionState == devices.Connected
}

func (d *Device) SetConnectionState(state devices.ConnectionState) {
	d.connectionState = state
}

func (d *Device) StartRecording(path string) error {
	// d.recordingSessionProcess = devices.NewCommand("xcrun", "simctl", "io", d.DeviceID(), "recordVideo", "—", "type=mp4", fmt.Sprintf("./%s.mp4", path))
	// if err := d.recordingSessionProcess.Start(); err != nil {
	// 	return err
	// }
	return nil
}

func (d *Device) StopRecording() error {
	var err error
	// if d.recordingSessionProcess != nil && d.recordingSessionProcess.Process != nil {
	// 	err = d.recordingSessionProcess.Process.Kill()
	// 	d.recordingSessionProcess = nil
	// }
	return err
}

func (d *Device) ConnectionTimeout() time.Duration {
	return CONNECTION_TIMEOUT
}
