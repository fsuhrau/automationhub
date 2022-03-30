package iosdevice

import (
	"bytes"
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/modules/webdriver"
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
	deviceModel             string
	deviceState             device.State
	deviceIP                net.IP
	recordingSessionProcess *exec.Cmd
	lastUpdateAt            time.Time
	webDriver               *webdriver.Client
}

func (d *Device) DeviceModel() string {
	return d.deviceModel
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

func (d *Device) UpdateDeviceInfos() error {
	return nil
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
		device, err := ios.GetDevice(d.DeviceID())
		if err != nil {
			return err
		}
		conn, err := zipconduit.New(device)
		if err != nil {
			return err
		}
		err = conn.SendFile(params.AppPath)
		return err
	*/
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	return nil
	/*
		device, err := ios.GetDevice(d.DeviceID())
		if err != nil {
			return err
		}
		svc, err := installationproxy.New(device)
		if err != nil {
			return err
		}
		err = svc.Uninstall(params.Identifier)
		return err
	*/
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {
	// d.StartXCUITestRunner()

	if d.webDriver == nil {
		return fmt.Errorf("webdriver not connected")
	}
	return d.webDriver.Launch(params.Identifier, true, []string{"SESSION_ID", sessionId, "DEVICE_ID", d.deviceID, "HOST", hostIP.String()})
}

func (d *Device) StopApp(params *app.Parameter) error {
	if d.webDriver == nil {
		return fmt.Errorf("webdriver nocht connected")
	}
	return d.webDriver.Terminate(params.Identifier)
}

func (d *Device) IsAppConnected() bool {
	return d.Connection() != nil
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
		cmd := exec2.NewCommand("idevicescreenshot", "-u", d.deviceID, fileName)
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

func (d *Device) RunNativeScript(script []byte) {

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
	time.Sleep(4 * time.Second)

	// d.webDriver = webdriver.New(fmt.Sprintf("http://%s:8100", d.deviceIP.String()))
	d.webDriver = webdriver.New("http://169.254.208.38:8100")
	d.webDriver.CreateSession()
	/*
		settings, err := d.webDriver.GetSettings()
		fmt.Println(settings)
	*/
	return nil
}

func (d *Device) StopXCUITestRunner() error {
	d.webDriver.CloseSession()
	_ = testmanagerd.CloseXCUITestRunner()
	d.webDriver = nil
	d.deviceState = device.StateRemoteDisconnected
	return nil
}
