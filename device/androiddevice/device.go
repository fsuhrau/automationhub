package androiddevice

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/tools/android"
	"github.com/spf13/viper"
	"image"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

var (
	DeviceLockRegex  = regexp.MustCompile(`.*mDreamingLockscreen=(true|false).*`)
	DeviceAwakeRegex = regexp.MustCompile(`.*mHoldingDisplaySuspendBlocker=(true|false).*`)
)

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
	testRecordingPath   string
	recordingSession    *exec.Cmd
	cfg                 *config.Device
	lastUpdateAt        time.Time
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
	case "RemoteDisconnected":
		d.deviceState = device.RemoteDisconnected
	case "Unknown":
		d.deviceState = device.Unknown
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
	isAwake, err := d.IsAwake()
	if err != nil {
		return err
	}
	if !isAwake {
		_ = d.pressKey(KEYCODE_WAKEUP)
		_ = d.pressKey(KEYCODE_MENU)
		_ = d.swipe(400, 800, 400, 200)
	}

	isLocked, err := d.IsLocked()
	if err != nil {
		return err
	}
	if isLocked {
		if d.cfg != nil && len(d.cfg.PIN) > 0 {
			for i := range d.cfg.PIN {
				offset := int(d.cfg.PIN[i] - '0')
				_ = d.pressKey(KEYCODE_NUMPAD_0 + offset)
			}
			_ = d.pressKey(KEYCODE_NUMPAD_ENTER)
		}
	}
	return nil
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {
	if err := d.unlockScreen(); err != nil {
		return err
	}
	if viper.GetBool("restart") {
		cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "am", "start", "-n", fmt.Sprintf("%s/%s", params.Identifier, params.LaunchActivity), "-e", "SESSION_ID", sessionId, "-e", "HOST", hostIP.String())
		return cmd.Run()
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	if viper.GetBool("restart") {
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

func (d *Device) IsLocked() (bool, error) {
	cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "dumpsys", "window", "|", "grep", "mDreamingLockscreen")
	out, err := cmd.Output()
	if err != nil {
		return true, err
	}
	findings := DeviceLockRegex.FindStringSubmatch(string(out))
	if len(findings) == 0 {
		return true, fmt.Errorf("IsLocked Regex cant be evaluated")
	}
	return findings[1] == "true", nil
}

func (d *Device) IsAwake() (bool, error) {
	cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "dumpsys", "power", "|", "grep", "mHoldingDisplaySuspendBlocker")
	out, err := cmd.Output()
	if err != nil {
		return true, err
	}
	findings := DeviceAwakeRegex.FindStringSubmatch(string(out))
	if len(findings) == 0 {
		return true, fmt.Errorf("IsAwake Regex cant be evaluated")
	}
	return findings[1] == "true", nil
}

func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	fileName := fmt.Sprintf("%s.png", d.deviceID)
	var width int
	var height int
	// cmd := device.NewCommand("adb", "-s", d.DeviceID(), "exec-out", "screencap", "-p", ">", fileName)
	if false {
		cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "screencap", "-p", "/sdcard/"+fileName)
		if err := cmd.Run(); err != nil {
			return nil, width, height, err
		}

		cmd = device.NewCommand("adb", "-s", d.DeviceID(), "pull", "/sdcard/"+fileName)
		if err := cmd.Run(); err != nil {
			return nil, width, height, err
		}

		cmd = device.NewCommand("adb", "-s", d.DeviceID(), "shell", "rm", "/sdcard/"+fileName)
		if err := cmd.Run(); err != nil {
			return nil, width, height, err
		}
	} else {
		start := time.Now()
		cmd := device.NewCommand("/bin/sh", "android_screen.sh", d.DeviceID(), fileName)
		if err := cmd.Run(); err != nil {
			return nil, width, height, err
		}
		logrus.Infof("Android Take Screenshot took: %d ms", time.Now().Sub(start).Milliseconds())
	}

	start := time.Now()
	imagePath, _ := os.Open(fileName)
	defer imagePath.Close()

	srcImage, _, _ := image.Decode(imagePath)

	width = srcImage.Bounds().Dx()
	height = srcImage.Bounds().Dy()
	imagePath.Seek(0, 0)
	bytes, err := ioutil.ReadAll(imagePath)
	logrus.Infof("Android open Screenshot took: %d ms", time.Now().Sub(start).Milliseconds())
	return bytes, width, height, err
}

func (d *Device) HasFeature(feature string) bool {
	if feature == "back" {
		return true
	}
	return false
}

func (d *Device) Execute(feature string) {
	features := map[string]func(d *Device){
		"back": func(d *Device) {
			d.pressKey(KEYCODE_BACK)
		},
	}
	if v, ok := features[feature]; ok {
		v(d)
	}
}

func (d *Device) Tap(x, y int64) error {
	cmd := device.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
	return cmd.Run()
}

func (d *Device) ConnectionTimeout() time.Duration {
	return CONNECTION_TIMEOUT
}
