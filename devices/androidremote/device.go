package androiddevice

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/devices"
	"github.com/sirupsen/logrus"
)

var restart bool = true

const CONNECTION_TIMEOUT = 40 * time.Second

type Device struct {
	deviceOSName    string
	deviceOSVersion string
	deviceName      string
	deviceID        string
	deviceState     devices.State
	deviceUSB       string
	product         string
	deviceModel     string
	transportID     string
	connectionState devices.ConnectionState
	launcActivity   string
	deviceIP        net.IP

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

func (d *Device) IsAppInstalled(bundleId string) bool {
	cmd := devices.NewCommand("adb", "-s", d.DeviceID(), "shell", "pm", "list", "packages")
	output, _ := cmd.Output()
	return strings.Contains(string(output), bundleId)
}

func (d *Device) getApiLevel() int64 {
	// pattern := regexp.MustCompile(`([0-9]{2,})`)

	cmd := devices.NewCommand("adb", "-s", d.DeviceID(), "shell", "getprop", "ro.build.version.sdk")
	output, err := cmd.Output()
	if err != nil {
		return -1
	}
	versionSring := string(output)
	versionSring = strings.Trim(versionSring, "\n")

	apiLevel, err := strconv.ParseInt(versionSring, 10, 64)
	if err != nil {
		logrus.Errorf("Could not parse: %v", err)
		return -1
	}

	return apiLevel
}

func (d *Device) findLaunchActivity(apkPath string) error {
	cmd := devices.NewCommand("aapt", "d", "badging", apkPath)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	regex := regexp.MustCompile(`launchable-activity:\s+name='([a-zA-Z0-9.]+)'\s+label='(.*)'\sicon='.*'`)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		matches := regex.FindAllStringSubmatch(line, -1)
		if len(matches) < 1 {
			continue
		}
		if len(matches[0]) < 3 {
			continue
		}
		d.launcActivity = matches[0][1]
		break
	}
	return nil
}

func (d *Device) lookupDeviceIP() error {
	cmd := devices.NewCommand("adb", "-s", d.DeviceID(), "shell", "ip", "-f", "inet", "addr", "show", "wlan0")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	regex := regexp.MustCompile(`\s+inet\s+([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})\/[0-9]+\sbrd\s([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})\s+.*`)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		matches := regex.FindAllStringSubmatch(line, -1)
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

func (d *Device) ExtractAppParameters(apkPath string) error {
	if err := d.findLaunchActivity(apkPath); err != nil {
		return err
	}

	if err := d.lookupDeviceIP(); err != nil {
		return err
	}

	return nil
}

func (d *Device) InstallApp(apkPath string) error {

	apiLevel := d.getApiLevel()
	isApkDebuggable := isDebuggablePackage(apkPath)

	var debug string
	if isApkDebuggable {
		debug = "d"
	}

	parameter := []string{"-s", d.DeviceID(), "install"}
	if apiLevel < 24 {
		parameter = append(parameter, []string{"-rg" + debug, apkPath}...)
	} else if isApkDebuggable {
		parameter = append(parameter, []string{"-r", "-g", "-d", apkPath}...)
	} else {
		parameter = append(parameter, []string{"-r", "-g", apkPath}...)
	}

	cmd := devices.NewCommand("adb", parameter...)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()

	logrus.Debugf("%v", output)

	if strings.Contains(string(output), "Failure [") {
		return fmt.Errorf("Installation Failed")
	}

	return err
}

func (d *Device) UninstallApp(bundleId string) error {
	cmd := devices.NewCommand("adb", "-s", d.DeviceID(), "uninstall", bundleId)
	return cmd.Run()
}

func (d *Device) unlockScreen() error {
	d.pressKey(KEYCODE_WAKEUP)
	d.pressKey(KEYCODE_MENU)
	d.swipe(400, 800, 400, 200)
	return nil
}

func (d *Device) StartApp(appPath string, bundleId string, sessionID string) error {
	d.unlockScreen()
	if restart {
		cmd := devices.NewCommand("adb", "-s", d.DeviceID(), "shell", "am", "start", "-n", fmt.Sprintf("%s/%s", bundleId, d.launcActivity))
		return cmd.Run()
	}
	return nil
}

func (d *Device) StopApp(appPath, bundleId string) error {
	if restart {
		cmd := devices.NewCommand("adb", "-s", d.DeviceID(), "shell", "am", "force-stop", bundleId)
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
	d.testRecordingPath = fmt.Sprintf("./%s.mp4", path)
	d.recordingSession = devices.NewCommand("adb", "-s", d.DeviceID(), "shell", "screenrecord", "--verbose", "/data/local/tmp/automation_hub_record.mp4")
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
	cmd := devices.NewCommand("adb", "-s", d.DeviceID(), "shell", "pull", "/data/local/tmp/automation_hub_record.mp4")
	if err := cmd.Run(); err != nil {
		return err
	}

	return os.Rename("automation_hub_record.mp4", d.testRecordingPath)
}

func (d *Device) ConnectionTimeout() time.Duration {
	return CONNECTION_TIMEOUT
}
