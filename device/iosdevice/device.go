package iosdevice

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config"
	"github.com/sirupsen/logrus"

	"github.com/disintegration/imaging"
	"github.com/fsuhrau/automationhub/device"
)

const (
	IOS_DEPLOY_BIN     = "ios-deploy" // "/usr/local/bin/ios-deploy"
	CONNECTION_TIMEOUT = 120 * time.Second
)

type Device struct {
	deviceOSName            string
	deviceOSVersion         string
	deviceName              string
	deviceID                string
	deviceState             device.State
	connectionState         device.ConnectionState
	deviceIP                net.IP
	recordingSessionProcess *exec.Cmd
	runningAppProcess       *exec.Cmd
	cfg                     *config.Device
	lastUpdateAt            time.Time
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

func (d *Device) IsAppInstalled(params *app.Parameter) (bool, error) {
	cmd := device.NewCommand(IOS_DEPLOY_BIN, "--id", d.DeviceID(), "--exists", "--bundle_id", params.Identifier)
	output, _ := cmd.Output()
	out := string(output)
	return strings.Contains(out, "true"), nil
}

func (d *Device) InstallApp(params *app.Parameter) error {
	cmd := device.NewCommand(IOS_DEPLOY_BIN, "--id", d.DeviceID(), "--bundle", params.AppPath)
	return cmd.Run()
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	cmd := device.NewCommand(IOS_DEPLOY_BIN, "--id", d.DeviceID(), "--uninstall_only", "--bundle_id", params.Identifier)
	return cmd.Run()
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {
	d.runningAppProcess = device.NewCommand(IOS_DEPLOY_BIN, "--json", "--id", d.DeviceID(), "--noinstall", "--noninteractive", "--no-wifi", "--bundle", params.AppPath, "--bundle_id", params.Identifier, "--args", fmt.Sprintf("SESSION_ID %s HOST %s", sessionId, hostIP.String()))
	if false {
		d.runningAppProcess.Stdout = os.Stdout
	}
	d.runningAppProcess.Stderr = os.Stderr
	if err := d.runningAppProcess.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	var err error
	if d.runningAppProcess != nil && d.runningAppProcess.Process != nil {
		err = d.runningAppProcess.Process.Signal(os.Interrupt)
		if err != nil {
			logrus.Errorf("Stop Interrupt error: %v", err)
		}
		//err = d.runningAppProcess.Process.Kill()
		//if err != nil {
		//	logrus.Errorf("Stop Kill error: %v", err)
		//}
		if e := d.runningAppProcess.Wait(); e != nil {
			logrus.Errorf("Stop Kill error: %v", e)
		}
		d.runningAppProcess = nil
	}
	return err
}

func (d *Device) IsAppConnected() bool {
	return d.connectionState == device.Connected
}

func (d *Device) SetConnectionState(state device.ConnectionState) {
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

func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	var width int
	var height int
	fileName := fmt.Sprintf("%s.png", d.deviceID)
	cmd := device.NewCommand("idevicescreenshot", "-u", d.deviceID, fileName)
	if err := cmd.Run(); err != nil {
		return nil, width, height, err
	}
	imagePath, _ := os.Open(fileName)
	defer imagePath.Close()
	srcImage, _, _ := image.Decode(imagePath)

	uploadedImage := imaging.Rotate(srcImage, -90, color.Gray{})
	width = uploadedImage.Bounds().Dx()
	height = uploadedImage.Bounds().Dy()
	var data []byte
	writer := bytes.NewBuffer(data)
	err := png.Encode(writer, uploadedImage)
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
