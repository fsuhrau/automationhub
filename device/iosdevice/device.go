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
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/screenshotr"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/disintegration/imaging"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

const (
	ConnectionTimeout = 120 * time.Second
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
	webDriver               *webdriver.Client
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
		if d.webDriver != nil {
			d.deviceState = device.StateBooted
		} else {
			d.deviceState = device.StateRemoteDisconnected
		}
	case "StateShutdown":
		d.deviceState = device.StateShutdown
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

	svc, _ := installationproxy.New(device)
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

func (d *Device) InstallApp(params *app.Parameter) error {
	return nil
	/*
		device, err := ios.GetDevice(d.deviceId())
		if err != nil {
			return err
		}
		conn, err := zipconduit.New(device)
		if err != nil {
			return err
		}
		err = conn.SendFile(params.appPath)
		return err
	*/
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	return nil
	/*
		device, err := ios.GetDevice(d.deviceId())
		if err != nil {
			return err
		}
		svc, err := installationproxy.New(device)
		if err != nil {
			return err
		}
		err = svc.Uninstall(params.identifier)
		return err
	*/
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
	fileName := fmt.Sprintf("%s.png", d.deviceID)

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

	err = ioutil.WriteFile(fileName, imageBytes, 0777)
	if err != nil {
		return nil, width, height, err
	}

	/*
		cmd := exec2.NewCommand("idevicescreenshot", "-u", d.deviceId, fileName)
		if err := cmd.Run(); err != nil {
			return nil, width, height, err
		}
	*/

	imagePath, _ := os.Open(fileName)
	defer imagePath.Close()
	srcImage, _, _ := image.Decode(imagePath)

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

func (d *Device) StartXCUITestRunner() error {
	device, err := ios.GetDevice(d.DeviceID())
	if err != nil {
		return err
	}

	bundleID, testbundleID, xctestconfig := "com.automationhub.WebDriverAgentRunner.xctrunner", "com.automationhub.WebDriverAgentRunner.xctrunner", "WebDriverAgentRunner.xctest"

	var wdaArg []string
	var wdaEnv []string

	go func() {
		err := testmanagerd.RunXCUIWithBundleIds(bundleID, testbundleID, xctestconfig, device, wdaArg, wdaEnv)
		fmt.Println(err)
	}()

	address, ok := <-webdriver.WDAHook.Connected
	if !ok {
		return fmt.Errorf("webdriver wdahook not connected")
	}
	d.webDriver = webdriver.New(address)
	d.webDriver.CreateSession()
	return nil
}

func (d *Device) StopXCUITestRunner() error {
	if d.webDriver != nil {
		d.webDriver.CloseSession()
	}
	_ = testmanagerd.CloseXCUITestRunner()
	d.webDriver = nil
	d.deviceState = device.StateRemoteDisconnected
	return nil
}

func (d *Device) RunNativeScript(script []byte) {
	scriptHandler := New(d.webDriver, d)
	if err := scriptHandler.Execute(string(script)); err != nil {
		d.Error("device", "Script Failed: %v", err)
		if d.webDriver != nil {
			d.webDriver.CloseSession()
		}
		d.webDriver.PressButton(KEYCODE_HOME)
	}
}
