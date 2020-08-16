package androiddevice

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/tools/android"
	"image"
	"image/png"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

var restart bool = true

const CONNECTION_TIMEOUT = 2 * time.Minute

type Device struct {
	deviceOSName        string
	deviceOSVersion     string
	deviceName          string
	deviceID            string
	deviceState         device.State
	deviceUSB           string
	product             string
	deviceModel         string
	transportID         string
	connectionState     device.ConnectionState
	deviceIP            net.IP
	deviceSupportedABIS []string
	deviceAPILevel      int64

	testRecordingPath string

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

func (d *Device) IsAppInstalled(params *app.Parameter) bool {
	return android.IsAppInstalled(d.deviceID, params)
}

func (d *Device) UpdateDeviceInfos() error {
	d.deviceAPILevel = android.GetParameterInt(d.deviceID, "ro.build.version.sdk")
	d.deviceOSVersion = android.GetParameterString(d.deviceID, "ro.build.version.release")
	d.deviceSupportedABIS = strings.Split(android.GetParameterString(d.deviceID, "ro.product.cpu.abilist"), ",")
	var err error
	d.deviceIP, err = android.GetDeviceIP(d.deviceID)
	return err
}

func (d *Device) InstallApp(params *app.Parameter) error {
	isApkDebuggable := isDebuggablePackage(params.AppPath)

	var debug string

	if isApkDebuggable {
		debug = "d"
	}
	parameter := []string{"-s", d.DeviceID(), "install"}
	if d.deviceAPILevel < 24 {
		parameter = append(parameter, []string{"-rg" + debug, params.AppPath}...)
		//} else if isApkDebuggable {
		//	parameter = append(parameter, []string{"-r", "-g", "-d", params.AppPath}...)
	} else {
		parameter = append(parameter, []string{"-r", "-g", params.AppPath}...)
	}

	cmd := device.NewCommand("adb", parameter...)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()

	logrus.Debugf("%v", output)

	if strings.Contains(string(output), "Failure [") {
		return fmt.Errorf("Installation Failed")
	}

	return err
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	cmd := device.NewCommand("adb", "-s", d.DeviceID(), "uninstall", params.Identifier)
	return cmd.Run()
}

func (d *Device) unlockScreen() error {
	d.pressKey(KEYCODE_WAKEUP)
	d.pressKey(KEYCODE_MENU)
	d.swipe(400, 800, 400, 200)
	return nil
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {
	d.unlockScreen()
	if restart {
		cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "am", "start", "-n", fmt.Sprintf("%s/%s", params.Identifier, params.LaunchActivity), "-e", "SESSION_ID", sessionId, "-e", "HOST", hostIP.String())
		return cmd.Run()
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	if restart {
		cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "am", "force-stop", params.Identifier)
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
	d.testRecordingPath = fmt.Sprintf("./%s.mp4", path)
	d.recordingSession = device.NewCommand("adb", "-s", d.DeviceID(), "shell", "screenrecord", "--verbose", "/data/local/tmp/automation_hub_record.mp4")
	if err := d.recordingSession.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Device) StopRecording() error {
	if d.recordingSession != nil && d.recordingSession.Process != nil {
		err := d.recordingSession.Process.Signal(os.Interrupt)
		d.recordingSession = nil
		if err != nil {
			return err
		}
	}
	cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "pull", "/data/local/tmp/automation_hub_record.mp4")
	if err := cmd.Run(); err != nil {
		return err
	}

	return os.Rename("automation_hub_record.mp4", d.testRecordingPath)
}

func (d *Device)GetScreenshot() ([]byte, error) {
	fileName := fmt.Sprintf("%s.png", d.deviceID)
	// cmd := device.NewCommand("adb", "-s", d.DeviceID(), "exec-out", "screencap", "-p", ">", fileName)
	if (false) {
		cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "screencap", "-p", "/sdcard/" + fileName)
		if err := cmd.Run(); err != nil {
			return nil, err
		}

		cmd = device.NewCommand("adb", "-s", d.DeviceID(), "pull", "/sdcard/" + fileName)
		if err := cmd.Run(); err != nil {
			return nil, err
		}

		cmd = device.NewCommand("adb", "-s", d.DeviceID(), "shell", "rm", "/sdcard/" + fileName)
		if err := cmd.Run(); err != nil {
			return nil, err
		}
	} else {
		cmd := device.NewCommand("/bin/sh", "android_screen.sh", d.DeviceID(), fileName)
		if err := cmd.Run(); err != nil {
			return nil, err
		}
	}

	imagePath, _ := os.Open(fileName)
	defer imagePath.Close()
	srcImage, _, _ := image.Decode(imagePath)

	width := float64(srcImage.Bounds().Dx())
	height := float64(srcImage.Bounds().Dy())
	srcImage.Bounds().Dy()
	factor := 640.0 / height
	resultImage := imaging.Resize(srcImage, int(width * factor), int(height * factor), imaging.Linear)
	var data []byte
	writer := bytes.NewBuffer(data)
	err := png.Encode(writer, resultImage)
	return writer.Bytes(), err
}

func (d *Device) HasFeature(feature string) bool {
	if feature == "back" {
		return true
	}
	return false
}

func (d *Device) Execute(feature string) {
	features := map[string] func(d *Device) {
		"back": func(d *Device){
			d.pressKey(KEYCODE_BACK)
		},
	}
	if v, ok := features[feature]; ok {
		v(d)
	}
}

func (d *Device) ConnectionTimeout() time.Duration {
	return CONNECTION_TIMEOUT
}
