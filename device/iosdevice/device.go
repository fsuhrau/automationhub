package iosdevice

import (
	"bytes"
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/modules/webdriver"
	"github.com/fsuhrau/automationhub/storage/models"
	"image"
	"image/color"
	"image/png"
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/screenshotr"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/danielpaulus/go-ios/ios/zipconduit"
	"github.com/disintegration/imaging"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

const (
	ConnectionTimeout = 120 * time.Second
	WDAStartTimeout   = 60 * time.Second
)

type Device struct {
	generic.Device
	deviceOSName            string
	deviceOSVersion         string
	deviceName              string
	deviceID                string
	deviceState             device.State
	deviceIP                net.IP
	recordingSessionProcess *exec.Cmd
	lastUpdateAt            time.Time
	wdaMu                   sync.Mutex // serializes starting and stopping WebDriverAgent
	webDriver               *webdriver.Client
	screenScale             float64 // screenshot pixels per WDA point, 0 = unknown
	deviceParameter         map[string]string
}

func (d *Device) DeviceParameter() map[string]string {
	return d.deviceParameter
}

func (d *Device) DeviceType() int {
	return int(models.DeviceTypePhone)
}
func (d *Device) PlatformType() int {
	return int(models.PlatformTypeiOS)
}

func (d *Device) DeviceOSName() string {
	return d.deviceOSName
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

func (d *Device) SetDeviceState(state string) {
	switch state {
	case "StateBooted":
		// attached and readable; WebDriverAgent is started on demand
		// (StartDevice), so it doesn't decide whether the device is ready
		d.deviceState = device.StateBooted
	case "StateShutdown":
		d.deviceState = device.StateShutdown
	case "StateRemoteDisconnected":
		d.deviceState = device.StateRemoteDisconnected
	case "StateUnauthorized":
		d.deviceState = device.StateUnauthorized
	default:
		d.deviceState = device.StateUnknown
	}
}

func (d *Device) UpdateDeviceInfos(response ios.GetAllValuesResponse) {
	d.deviceName = response.Value.DeviceName
	d.deviceOSName = response.Value.ProductName
	d.deviceOSVersion = response.Value.ProductVersion

	d.deviceParameter = make(map[string]string)
	d.deviceParameter["Device Model"] = response.Value.ProductType
}

func (d *Device) IsAppInstalled(params *app.Parameter) (bool, error) {
	device, err := ios.GetDevice(d.DeviceID())
	if err != nil {
		return false, err
	}

	svc, err := installationproxy.New(device)
	if err != nil {
		return false, fmt.Errorf("installation proxy: %w", err)
	}
	defer svc.Close()

	response, err := svc.BrowseUserApps()
	if err != nil {
		return false, err
	}

	for _, i := range response {
		if i.CFBundleIdentifier == params.Identifier {
			return true, nil
		}
	}

	return false, nil
}

// InstallApp installs an .ipa (or .app directory) through the zipconduit service.
func (d *Device) InstallApp(params *app.Parameter) error {
	if params == nil || params.App == nil || params.App.AppPath == "" {
		return fmt.Errorf("install app: no app path given")
	}
	device, err := ios.GetDevice(d.DeviceID())
	if err != nil {
		return err
	}
	conn, err := zipconduit.New(device)
	if err != nil {
		return fmt.Errorf("install app: zipconduit: %w", err)
	}
	defer conn.Close()

	d.Log("device", "Install App '%s' from %s", params.Identifier, params.App.AppPath)
	if err := conn.SendFile(params.App.AppPath); err != nil {
		return fmt.Errorf("install app %s: %w", params.Identifier, err)
	}
	return nil
}

// UninstallApp removes the app with the given bundle identifier.
func (d *Device) UninstallApp(params *app.Parameter) error {
	if params == nil || params.Identifier == "" {
		return fmt.Errorf("uninstall app: no bundle identifier given")
	}
	device, err := ios.GetDevice(d.DeviceID())
	if err != nil {
		return err
	}
	svc, err := installationproxy.New(device)
	if err != nil {
		return fmt.Errorf("uninstall app: installation proxy: %w", err)
	}
	defer svc.Close()

	d.Log("device", "Uninstall App '%s'", params.Identifier)
	if err := svc.Uninstall(params.Identifier); err != nil {
		return fmt.Errorf("uninstall app %s: %w", params.Identifier, err)
	}
	return nil
}

func (d *Device) StartApp(_ *device.DeviceConfig, appParams *app.Parameter, sessionId string, nodeUrl string) error {
	if d.webDriver == nil {
		return fmt.Errorf("webdriver not connected")
	}
	return d.webDriver.Launch(appParams.Identifier, true, []string{"SESSION_ID", sessionId, "DEVICE_ID", d.deviceID, "NODE_URL", nodeUrl})
}

func (d *Device) StopApp(params *app.Parameter) error {
	if d.webDriver == nil {
		return fmt.Errorf("webdriver nocht connected")
	}
	info, err := d.webDriver.ActiveAppInfo()
	if err != nil {
		return err
	}
	if info.Value.BundleId == params.Identifier {
		return d.webDriver.Terminate(params.Identifier)
	}
	return nil
}

func (d *Device) IsAppConnected() bool {
	return d.Connection() != nil
}

func (d *Device) StartRecording(path string) error {
	// d.recordingSessionProcess = devices.NewCommand("xcrun", "simctl", "io", d.deviceId(), "recordVideo", "—", "type=mp4", fmt.Sprintf("./%s.mp4", path))
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
	device, err := ios.GetDevice(d.DeviceID())
	if err != nil {
		return nil, width, height, err
	}

	screenshotrService, err := screenshotr.New(device)
	if err != nil {
		return nil, width, height, err
	}
	imageBytes, err := screenshotrService.TakeScreenshot()
	if err != nil {
		return nil, width, height, err
	}

	srcImage, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, width, height, fmt.Errorf("decode screenshot: %w", err)
	}

	uploadedImage := imaging.Rotate(srcImage, -90, color.Gray{})
	width = uploadedImage.Bounds().Dx()
	height = uploadedImage.Bounds().Dy()
	var data []byte
	writer := bytes.NewBuffer(data)
	err = png.Encode(writer, uploadedImage)
	return writer.Bytes(), width, height, err
}

func (d *Device) HasFeature(string) bool {
	return false
}

func (d *Device) Execute(string) {

}

func (d *Device) ConnectionTimeout() time.Duration {
	return ConnectionTimeout
}

// StartXCUITestRunner starts WebDriverAgent and creates a session. A runner
// that is already up (e.g. started for the live video) is reused.
func (d *Device) StartXCUITestRunner() error {
	d.wdaMu.Lock()
	defer d.wdaMu.Unlock()
	return d.startWebDriver()
}

// ensureWebDriver starts WebDriverAgent unless it is up already.
func (d *Device) ensureWebDriver() (*webdriver.Client, error) {
	d.wdaMu.Lock()
	defer d.wdaMu.Unlock()
	if wd := d.webDriver; wd != nil && wd.Alive(3*time.Second) {
		return wd, nil
	}
	if err := d.startWebDriver(); err != nil {
		return nil, err
	}
	return d.webDriver, nil
}

func (d *Device) startWebDriver() error {
	if wd := d.webDriver; wd != nil && wd.Alive(3*time.Second) {
		if _, err := wd.CreateSession(); err != nil {
			return fmt.Errorf("create WebDriverAgent session: %w", err)
		}
		return nil
	}
	device, err := ios.GetDevice(d.DeviceID())
	if err != nil {
		return err
	}

	bundleID, testbundleID, xctestconfig := "com.automationhub.WebDriverAgentRunner.xctrunner", "com.automationhub.WebDriverAgentRunner.xctrunner", "WebDriverAgentRunner.xctest"

	var wdaArg []string
	var wdaEnv []string

	address, err := webdriver.WDAHook.Await(WDAStartTimeout, func() {
		go func() {
			if err := testmanagerd.RunXCUIWithBundleIds(bundleID, testbundleID, xctestconfig, device, wdaArg, wdaEnv); err != nil {
				d.Error("device", "WebDriverAgent runner stopped: %v", err)
			}
		}()
	})
	if err != nil {
		return err
	}
	d.webDriver = webdriver.New(address)
	d.screenScale = 0
	if _, err := d.webDriver.CreateSession(); err != nil {
		return fmt.Errorf("create WebDriverAgent session: %w", err)
	}
	return nil
}

func (d *Device) StopXCUITestRunner() error {
	d.wdaMu.Lock()
	defer d.wdaMu.Unlock()
	if d.webDriver != nil {
		d.webDriver.CloseSession()
	}
	_ = testmanagerd.CloseXCUITestRunner()
	d.webDriver = nil
	// the device stays attached; its state follows usbmuxd, not WebDriverAgent
	return nil
}

func (d *Device) RunNativeScript(script []byte) {
	if d.webDriver == nil {
		d.Error("device", "Script Failed: no WebDriverAgent session for device %s", d.DeviceID())
		return
	}
	scriptHandler := New(d.webDriver, d)
	if err := scriptHandler.Execute(string(script)); err != nil {
		d.Error("device", "Script Failed: %v", err)
		d.webDriver.PressButton(KEYCODE_HOME)
		d.webDriver.CloseSession()
	}
}
