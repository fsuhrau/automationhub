package iosdevice

import (
	"bytes"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/sirupsen/logrus"
	"image"
	"image/color"
	"image/png"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fsuhrau/automationhub/device"
)

var reinstall bool = true

const CONNECTION_TIMEOUT = 120 * time.Second

type Device struct {
	deviceOSName    string
	deviceOSVersion string
	deviceName      string
	deviceID        string
	deviceState     device.State
	connectionState device.ConnectionState
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
	cmd := device.NewCommand("/usr/local/bin/ios-deploy", "--id", d.DeviceID(), "--exists", "--bundle_id", params.Identifier)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "true")
}

func (d *Device) InstallApp(params *app.Parameter) error {
	cmd := device.NewCommand("/usr/local/bin/ios-deploy", "--id", d.DeviceID(), "--bundle", params.AppPath)
	return cmd.Run()
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	cmd := device.NewCommand("/usr/local/bin/ios-deploy", "--id", d.DeviceID(), "--uninstall_only", "--bundle_id", params.Identifier)
	return cmd.Run()
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {
	if reinstall {
		d.runningAppProcess = device.NewCommand("/usr/local/bin/ios-deploy", "--json", "--id", d.DeviceID(), "--noinstall", "--noninteractive", "--no-wifi", "--bundle", params.AppPath, "--bundle_id", params.Identifier, "--args", fmt.Sprintf("SESSION_ID %s HOST %s", sessionId, hostIP.String()))
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

func (d *Device) StopApp(params *app.Parameter) error {
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

func (d *Device)GetScreenshot() ([]byte, error) {
	fileName := fmt.Sprintf("%s.png", d.deviceID)
	cmd := device.NewCommand("idevicescreenshot", "-u", d.deviceID, fileName)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	imagePath, _ := os.Open(fileName)
	defer imagePath.Close()
	srcImage, _, _ := image.Decode(imagePath)

	//srcDim := srcImage.Bounds()
	// dstImage := image.NewRGBA(image.Rect(0, 0, srcDim.Dy(), srcDim.Dx()))
	uploadedImage := imaging.Rotate(srcImage, -90, color.Gray{})
	// graphics.Rotate(dstImage, srcImage, &graphics.RotateOptions{math.Pi / 2.0})
	width := float64(srcImage.Bounds().Dy())
	height := float64(srcImage.Bounds().Dx())
	srcImage.Bounds().Dy()
	factor := 640.0 / height
	resultImage := imaging.Resize(uploadedImage, int(width * factor), int(height * factor), imaging.Linear)
	var data []byte
	writer := bytes.NewBuffer(data)
	err := png.Encode(writer, resultImage)
	return writer.Bytes(), err
	//return ioutil.ReadFile(fileName)
}

func (d *Device) HasFeature(string) bool {
	return false
}

func (d *Device) Execute(string) {

}

func (d *Device) ConnectionTimeout() time.Duration {
	return CONNECTION_TIMEOUT
}
