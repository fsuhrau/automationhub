package androiddevice

import (
	"bytes"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/storage/models"
	exec2 "github.com/fsuhrau/automationhub/tools/exec"
	"github.com/google/uuid"
	"image"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/tools/android"
	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"
)

var (
	DeviceLockRegex  = regexp.MustCompile(`.*mDreamingLockscreen=(true|false).*`)
	DeviceAwakeRegex = regexp.MustCompile(`.*mHoldingDisplaySuspendBlocker=(true|false).*`)
)

const ConnectionTimeout = 2 * time.Minute

type Device struct {
	generic.Device
	deviceOSName      string
	deviceOSVersion   string
	deviceName        string
	deviceID          string
	deviceState       device.State
	deviceUSB         string
	product           string
	deviceIP          net.IP
	deviceAPILevel    int64
	testRecordingPath string
	recordingSession  *exec.Cmd
	lastUpdateAt      time.Time
	installedApps     map[string]string
	deviceParameter   map[string]string
}

func (d *Device) DeviceOSName() string {
	return d.deviceOSName
}

func (d *Device) DeviceType() int {
	return int(models.DeviceTypePhone)
}

func (d *Device) PlatformType() int {
	return int(models.PlatformTypeAndroid)
}

func (d *Device) DeviceParameter() map[string]string {
	return d.deviceParameter
}

func (d *Device) DeviceOSVersion() string {
	return d.deviceOSVersion
}

func (d *Device) TargetVersion() string {
	return ""
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

func GetConnectionString(parameter *models.ConnectionParameter) string {
	return fmt.Sprintf("%s:%d", parameter.IP, parameter.Port)
}

func (d *Device) SetDeviceState(state string) {
	switch state {
	case "StateBooted":
		d.deviceState = device.StateBooted
	case "StateShutdown":
		d.deviceState = device.StateShutdown
	case "StateRemoteDisconnected":
		d.deviceState = device.StateRemoteDisconnected
	case "StateUnknown":
		d.deviceState = device.StateUnknown
	default:
		d.deviceState = device.StateUnknown
	}
}

func (d *Device) IsAppInstalled(params *app.Parameter) (bool, error) {
	isAppInstalled := false
	hash, ok := d.installedApps[params.Identifier]
	if ok {
		isAppInstalled = hash == params.App.Hash
	}
	installed, err := android.IsAppInstalled(d.deviceID, params)
	d.Log("device", "App '%s' is installed: %t", params.Identifier, installed)
	return installed && isAppInstalled, err
}

func (d *Device) UpdateDeviceInfos(infos []string) {
	init := false
	if d.deviceParameter == nil {
		d.deviceParameter = make(map[string]string)
		init = true
	}
	// deviceID := infos[1]
	deviceUSB := infos[2]
	product := infos[3]
	model := infos[4]
	name := infos[5]
	transportID := infos[6]

	d.deviceName = name
	d.product = product

	d.deviceParameter["Device Model"] = model
	d.deviceParameter["Transport ID"] = transportID
	d.deviceParameter["Device USB"] = deviceUSB

	if init {
		d.deviceAPILevel = android.GetParameterInt(d.deviceID, "ro.build.version.sdk")
		d.deviceOSVersion = android.GetParameterString(d.deviceID, "ro.build.version.release")
		d.deviceIP, _ = android.GetDeviceIP(d.deviceID)
		d.deviceParameter["IP"] = d.deviceIP.String()
		d.deviceParameter["ABIs"] = android.GetParameterString(d.deviceID, "ro.product.cpu.abilist")
		d.deviceParameter["API Level"] = fmt.Sprintf("%d", d.deviceAPILevel)
	}
}

func (d *Device) InstallApp(params *app.Parameter) error {

	d.Log("device", "Install App '%s'", params.Identifier)

	isApkDebuggable := isDebuggablePackage(params.App.AppPath)

	var debug string

	if isApkDebuggable {
		debug = "d"
	}
	parameter := []string{"-s", d.DeviceID(), "install", "-t"}
	if d.deviceAPILevel < 24 {
		parameter = append(parameter, []string{"-rg" + debug, params.App.AppPath}...)
		//} else if isApkDebuggable {
		//	parameter = append(parameter, []string{"-r", "-g", "-d", params.AppPath}...)
	} else {
		parameter = append(parameter, []string{"-r", "-g", params.App.AppPath}...)
	}

	cmd := exec2.NewCommand("adb", parameter...)
	//cmd.Stdout = os.Stdout
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation failed: \"%s\" %v", errb.String(), err)
	}

	d.installedApps[params.Identifier] = params.App.Hash

	return nil
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	d.Log("device", "Uninstall App '%s'", params.Identifier)

	cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "uninstall", params.Identifier)
	if err := cmd.Run(); err != nil {
		return err
	}
	d.installedApps[params.Identifier] = ""
	return nil
}

func (d *Device) unlockScreen() error {
	d.Log("device", "Unlock Screen")

	isAwake, err := d.IsAwake()
	if err != nil {
		return err
	}
	if !isAwake {
		_ = d.pressKey(KEYCODE_WAKEUP)
		_ = d.pressKey(KEYCODE_MENU)
		_ = d.swipe(400, 800, 400, 200)
	}

	isLocked, err := d.IsPinLocked()
	if err != nil {
		return err
	}
	if isLocked {
		if pin := d.GetConfig().GetAttribute(generic.AttributePin); len(pin) > 0 {
			for i := range pin {
				offset := int(pin[i] - '0')
				_ = d.pressKey(KEYCODE_NUMPAD_0 + offset)
			}
			_ = d.pressKey(KEYCODE_NUMPAD_ENTER)
		}
	}
	return nil
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, nodeUrl string) error {
	d.Log("device", "Start App '%s' with Session: '%s'", params.Identifier, sessionId)

	if err := d.unlockScreen(); err != nil {
		return err
	}
	cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "am", "start", "-n", fmt.Sprintf("%s/%s", params.Identifier, params.App.Android.LaunchActivity), "-e", "SESSION_ID", sessionId, "-e", "NODE_URL", nodeUrl, "-e", "DEVICE_ID", d.deviceID)
	return cmd.Run()
}

func (d *Device) StopApp(params *app.Parameter) error {
	d.Log("device", "Stop App '%s'", params.Identifier)
	if false {
		cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "am", "force-stop", params.Identifier)
		return cmd.Run()
	} else {
		cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "pm", "clear", params.Identifier)
		// adb shell pm clear com.my.app.package
		return cmd.Run()
	}

	return nil
}

func (d *Device) IsAppConnected() bool {
	return d.Connection() != nil
}

func (d *Device) StartRecording(path string) error {
	d.Log("device", "Start Recording Session")
	d.testRecordingPath = fmt.Sprintf("./%s.mp4", path)
	d.recordingSession = exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "screenrecord", "--verbose", "/data/local/tmp/automation_hub_record.mp4")
	if err := d.recordingSession.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Device) StopRecording() error {
	d.Log("device", "Stop Recording Session")
	if d.recordingSession != nil && d.recordingSession.Process != nil {
		if err := d.recordingSession.Process.Signal(os.Interrupt); err != nil {
			return err
		}
		if err := d.recordingSession.Wait(); err != nil {
			log.Infof("stop recording session failed: %v", err)
		}
		d.recordingSession = nil
	}
	time.Sleep(500 * time.Millisecond)
	cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "pull", "/data/local/tmp/automation_hub_record.mp4")
	if err := cmd.Run(); err != nil {
		return err
	}

	return os.Rename("automation_hub_record.mp4", d.testRecordingPath)
}

func (d *Device) IsPinLocked() (bool, error) {
	cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "dumpsys", "window", "|", "grep", "mDreamingLockscreen")
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
	cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "dumpsys", "power", "|", "grep", "mHoldingDisplaySuspendBlocker")
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
	d.Log("device", "Get Native Screenshot")
	id := uuid.New()
	fileName := fmt.Sprintf("%s.png", id.String())
	var width int
	var height int
	if true {
		cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "screencap", "-p", "/sdcard/"+fileName)
		if err := cmd.Run(); err != nil {
			return nil, width, height, err
		}

		cmd = exec2.NewCommand("adb", "-s", d.DeviceID(), "pull", "/sdcard/"+fileName)
		if err := cmd.Run(); err != nil {
			return nil, width, height, err
		}

		cmd = exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "rm", "/sdcard/"+fileName)
		if err := cmd.Run(); err != nil {
			return nil, width, height, err
		}
	} else {
		start := time.Now()
		cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "exec-out", "screencap", "-p", ">", fileName)
		//cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "screencap", "-p", ">", fileName)
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
	bytes, err := io.ReadAll(imagePath)
	logrus.Infof("Android open Screenshot took: %d ms", time.Now().Sub(start).Milliseconds())
	defer os.Remove(fileName)
	return bytes, width, height, err
}

func (d *Device) HasFeature(feature string) bool {
	if feature == "back" {
		return true
	}
	return false
}

func (d *Device) Execute(feature string) {
	d.Log("device", "Execute Feature: '%s'", feature)
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
	d.Log("device", "Execute Tap: %d,%d", x, y)
	cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
	return cmd.Run()
}

func (d *Device) ConnectionTimeout() time.Duration {
	return ConnectionTimeout
}

func (d *Device) getScreenXml() (*xmlquery.Node, error) {
	start := time.Now()
	defer func(s time.Time) {
		logrus.Infof("Android getScreenXml took: %d ms", time.Now().Sub(s).Milliseconds())
	}(start)

	cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "exec-out", "uiautomator", "dump", "/dev/tty")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(out)
	return xmlquery.Parse(reader)
}

func (d *Device) RunNativeScript(script []byte) {
	scriptHandler := New(d)
	if err := scriptHandler.Execute(string(script)); err != nil {
		d.Error("device", "Script Failed: %v", err)
		d.pressKey(KEYCODE_BACK)
	}
}
