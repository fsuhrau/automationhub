package iosdevice

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/screenshotr"
	"github.com/danielpaulus/go-ios/ios/zipconduit"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/modules/webdriver"
	"github.com/fsuhrau/automationhub/storage/models"
)

const (
	ConnectionTimeout = 120 * time.Second
	WDAStartTimeout   = 60 * time.Second
)

// Device is an iOS device attached over usbmux. The Info holds its identity,
// WebDriverAgent gives the app access.
type Device struct {
	generic.Device
	generic.Info
	wda          *launcher
	lastUpdateAt time.Time
	wdaMu        sync.Mutex // serializes starting and stopping WebDriverAgent
	webDriver    *webdriver.Client
	wdaRunner    *wdaRunner // set when xcodebuild runs WebDriverAgent
	screenScale  float64    // screenshot pixels per WDA point, 0 = unknown
}

// newDevice creates a device that starts WebDriverAgent with cfg.
func newDevice(id string, cfg *config.WebDriver, deps Deps) *Device {
	d := &Device{}
	d.wda = &launcher{cfg: cfg, deps: deps, logf: d.logf}
	d.ID = id
	d.Type = int(models.DeviceTypePhone)
	d.Platform = int(models.PlatformTypeiOS)
	return d
}

// UpdateDeviceInfos takes name, OS and model from the lockdown values.
func (d *Device) UpdateDeviceInfos(response ios.GetAllValuesResponse) {
	d.Name = response.Value.DeviceName
	d.OSName = response.Value.ProductName
	d.OSVersion = response.Value.ProductVersion
	d.Parameter = map[string]string{"Device Model": response.Value.ProductType}
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
	//nolint:errcheck // the transfer is over; there is nothing to do about a failed close
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

	uploadedImage := rotateClockwise(srcImage)
	width = uploadedImage.Bounds().Dx()
	height = uploadedImage.Bounds().Dy()
	var data []byte
	writer := bytes.NewBuffer(data)
	err = png.Encode(writer, uploadedImage)
	return writer.Bytes(), width, height, err
}

// rotateClockwise turns src by 90 degrees clockwise.
func rotateClockwise(src image.Image) *image.NRGBA {
	b := src.Bounds()
	in := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(in, in.Bounds(), src, b.Min, draw.Src)
	w, h := b.Dx(), b.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			si := in.PixOffset(x, y)
			di := out.PixOffset(h-1-y, x)
			copy(out.Pix[di:di+4], in.Pix[si:si+4])
		}
	}
	return out
}
