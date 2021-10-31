package androiddevice

import (
	"bytes"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/device/generic"
	exec2 "github.com/fsuhrau/automationhub/tools/exec"
	"image"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/tools/android"
	"github.com/prometheus/common/log"
	"github.com/spf13/viper"

	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

var (
	DeviceLockRegex  = regexp.MustCompile(`.*mDreamingLockscreen=(true|false).*`)
	DeviceAwakeRegex = regexp.MustCompile(`.*mHoldingDisplaySuspendBlocker=(true|false).*`)
)

const CONNECTION_TIMEOUT = 2 * time.Minute

type Device struct {
	generic.Device
	deviceOSName        string
	deviceOSVersion     string
	deviceName          string
	deviceID            string
	deviceState         device.State
	deviceUSB           string
	product             string
	deviceModel         string
	transportID         string
	deviceIP            net.IP
	deviceSupportedABIS []string
	deviceAPILevel      int64
	testRecordingPath   string
	recordingSession    *exec.Cmd
	cfg                 *config.Device
	lastUpdateAt        time.Time
	installedApps       map[string]string
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
		isAppInstalled = hash == params.Hash
	}
	installed, err := android.IsAppInstalled(d.deviceID, params)
	d.Log("device","App '%s' is installed: %t", params.Identifier, installed)
	return installed && isAppInstalled, err
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

	d.Log("device","Install App '%s'", params.Identifier)

	isApkDebuggable := isDebuggablePackage(params.AppPath)

	var debug string

	if isApkDebuggable {
		debug = "d"
	}
	parameter := []string{"-s", d.DeviceID(), "install", "-t"}
	if d.deviceAPILevel < 24 {
		parameter = append(parameter, []string{"-rg" + debug, params.AppPath}...)
		//} else if isApkDebuggable {
		//	parameter = append(parameter, []string{"-r", "-g", "-d", params.AppPath}...)
	} else {
		parameter = append(parameter, []string{"-r", "-g", params.AppPath}...)
	}

	cmd := exec2.NewCommand("adb", parameter...)
	//cmd.Stdout = os.Stdout
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation failed: \"%s\" %v", errb.String(), err)
	}

	d.installedApps[params.Identifier] = params.Hash

	return nil
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	d.Log("device","Uninstall App '%s'", params.Identifier)

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
	d.Log("device","Start App '%s' with Session: '%s'", params.Identifier, sessionId)

	if err := d.unlockScreen(); err != nil {
		return err
	}
	if viper.GetBool("restart") {
		cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "am", "start", "-n", fmt.Sprintf("%s/%s", params.Identifier, params.LaunchActivity), "-e", "SESSION_ID", sessionId, "-e", "HOST", hostIP.String(), "-e", "DEVICE_ID", d.deviceID)
		return cmd.Run()
	}
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	d.Log("device","Stop App '%s'", params.Identifier)

	if viper.GetBool("restart") {
		cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "am", "force-stop", params.Identifier)
		return cmd.Run()
	}
	return nil
}

func (d *Device) IsAppConnected() bool {
	return d.Connection() != nil
}

func (d *Device) StartRecording(path string) error {
	d.Log("device","Start Recording Session")
	d.testRecordingPath = fmt.Sprintf("./%s.mp4", path)
	d.recordingSession = exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "screenrecord", "--verbose", "/data/local/tmp/automation_hub_record.mp4")
	if err := d.recordingSession.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Device) StopRecording() error {
	d.Log("device","Stop Recording Session")
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
	d.Log("device","Get Native Screenshot")
	fileName := fmt.Sprintf("%s.png", d.deviceID)
	var width int
	var height int
	// cmd := device.NewCommand("adb", "-s", d.DeviceID(), "exec-out", "screencap", "-p", ">", fileName)
	if false {
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
		cmd := exec2.NewCommand("/bin/sh", "android_screen.sh", d.DeviceID(), fileName)
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
	d.Log("device","Execute Feature: '%s'", feature)
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
	d.Log("device","Execute Tap: %d,%d", x, y)
	cmd := exec2.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
	return cmd.Run()
}

func (d *Device) ConnectionTimeout() time.Duration {
	return CONNECTION_TIMEOUT
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

func (d *Device) RunNativeScript(script []byte)  {

	boundsEx := regexp.MustCompile(`\[([0-9]+),([0-9]+)\]\[([0-9]+),([0-9]+)\]`)

	actions, err := ParseNativeScript(script)
	if err != nil {
		d.Error("device", "Error: %v", err)
		d.pressKey(KEYCODE_BACK)
		return
	}

	_ = actions

	for _, a := range actions {
		if wfa, ok  := a.(*WaitForAction); ok {
			d.Log("device", "wait_for_action: %s with timeout: %d", wfa.XPath, wfa.Timeout)
			timeout := time.Now().Add(time.Duration(wfa.Timeout) * time.Second)

			for {
				xml, _ := d.getScreenXml()
				element := xmlquery.FindOne(xml, wfa.XPath)
				if element == nil {
					if time.Now().After(timeout) {
						d.Error("device", "Timeout: Element '%s' not found", wfa.XPath)
						d.pressKey(KEYCODE_BACK)
						return
					}
					time.Sleep(500 * time.Millisecond)
				} else {
					break
				}
			}
		}
		if ca, ok  := a.(*ClickAction); ok {
			d.Log("device", "click: %s", ca.XPath)
			xml, _ := d.getScreenXml()
			element := xmlquery.FindOne(xml, ca.XPath)
			if element == nil {
				d.Error("device", "Element '%s' not found", ca.XPath)
				d.pressKey(KEYCODE_BACK)
				return
			}

			var bounds string
			for _, attr  := range element.Attr {
				if attr.Name.Local == "bounds" {
					bounds = attr.Value
					break
				}
			}

			actionContent := boundsEx.FindAllStringSubmatch(bounds, -1)
			if len(actionContent) == 0 {
				d.Error("device", "No valid bounds for element '%s'", ca.XPath)
				d.pressKey(KEYCODE_BACK)
				return
			}
			xs, _ := strconv.ParseFloat(actionContent[0][1], 64)
			ys, _ := strconv.ParseFloat(actionContent[0][2], 64)
			xe, _ := strconv.ParseFloat(actionContent[0][3], 64)
			ye, _ := strconv.ParseFloat(actionContent[0][4], 64)

			x := (xs + xe) * 0.5
			y := (ys + ye) * 0.5

			if err := d.Tap(int64(x), int64(y)); err != nil {
				d.Error("device", "Touch element '%s' failed: '%v'", ca.XPath, err)
				d.pressKey(KEYCODE_BACK)
				return
			}
		}
	}
}