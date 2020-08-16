package androiddevice

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

var restart bool = true

var IPLookupRegex = regexp.MustCompile(`\s+inet\s+([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})/[0-9]+\sbrd\s([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})\s+.*`)

const CONNECTION_TIMEOUT = 40 * time.Second

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
	cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "pm", "list", "packages")
	output, _ := cmd.Output()
	return strings.Contains(string(output), params.Identifier)
}

func (d *Device) getParameterRaw(param string) string {
	cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "getprop", param)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return  strings.Trim(string(output), "\n")
}

func (d *Device) getParameterInt(param string) int64 {
	versionString := d.getParameterRaw(param)
	intParam, err := strconv.ParseInt(versionString, 10, 64)
	if err != nil {
		logrus.Errorf("Could not parse: %v", err)
		return -1
	}

	return intParam
}

func (d *Device) lookupDeviceIP() error {
	cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "ip", "-f", "inet", "addr", "show", "wlan0")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		matches := IPLookupRegex.FindAllStringSubmatch(line, -1)
		if len(matches) < 1 {
			continue
		}
		if len(matches[0]) < 3 {
			continue
		}
		d.deviceIP = net.ParseIP(matches[0][1])
		break
	}
	return nil
}

func (d *Device) UpdateParameter() error {
	d.deviceAPILevel = d.getParameterInt("ro.build.version.sdk")
	d.deviceOSVersion = d.getParameterRaw("ro.build.version.release")
	d.deviceSupportedABIS = strings.Split(d.getParameterRaw("ro.product.cpu.abilist"), ",")

	if err := d.lookupDeviceIP(); err != nil {
		return err
	}
	return nil
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
	} else if isApkDebuggable {
		parameter = append(parameter, []string{"-r", "-g", "-d", params.AppPath}...)
	} else {
		parameter = append(parameter, []string{"-r", "-g", params.AppPath}...)
	}

	cmd := device.NewCommand("adb", parameter...)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()

	logrus.Debugf("%v", output)

	if strings.Contains(string(output), "Failure [") {
		return fmt.Errorf("installation failed")
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

func (d *Device) StartApp(params *app.Parameter, sessionID string) error {
	d.unlockScreen()
	if restart {
		cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "am", "start", "-n", fmt.Sprintf("%s/%s", params.Identifier, params.LaunchActivity))
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
	fileName := fmt.Sprintf("tmp/%s.png", d.deviceID)
	cmd := device.NewCommand("adb", "-s", d.DeviceID(), "exec-out", "screencap", "-p", ">", fileName)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return ioutil.ReadFile(fileName)
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
