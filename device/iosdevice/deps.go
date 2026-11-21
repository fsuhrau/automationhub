package iosdevice

import (
	"net"
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/modules/webdriver"
)

// Forwarder is a local port that forwards to a port of the device.
type Forwarder interface {
	Addr() string
	Close()
}

// USBOps are the go-ios calls of the iOS packages as function values, so tests
// replace them. A nil field gets the go-ios implementation.
type USBOps struct {
	ListDevices func() (ios.DeviceList, error)
	GetValues   func(ios.DeviceEntry) (ios.GetAllValuesResponse, error)
	GetDevice   func(udid string) (ios.DeviceEntry, error)
	// Probe fails when nothing listens on the port of the device.
	Probe   func(entry ios.DeviceEntry, port uint16) error
	Forward func(entry ios.DeviceEntry, port uint16) (Forwarder, error)
	// RunXCUI runs a test runner on the device and blocks until it ends.
	RunXCUI   func(bundleID, testRunnerBundleID, xctestConfig string, entry ios.DeviceEntry, args, env []string) error
	CloseXCUI func() error
}

// Deps are the collaborators of the handler and its devices. A zero field gets
// its default, see withDefaults.
type Deps struct {
	Clock device.Clock
	HTTP  *http.Client
	// Runner starts xcodebuild; its output carries the WebDriverAgent URL, so
	// it is started with StartStreaming.
	Runner device.CommandRunner
	USB    USBOps
	// Hook picks the WebDriverAgent URL out of the go-ios log output. It only
	// sees the logs once it is added to the logger, see NewHandler.
	Hook *webdriver.WDAConnectHook

	openListener func() (muxListener, error)
}

// withDefaults fills the missing collaborators.
func (d Deps) withDefaults() Deps {
	if d.Clock == nil {
		d.Clock = device.RealClock{}
	}
	if d.HTTP == nil {
		d.HTTP = &http.Client{}
	}
	if d.Runner == nil {
		d.Runner = device.ExecRunner{}
	}
	if d.Hook == nil {
		d.Hook = &webdriver.WDAConnectHook{}
	}
	if d.openListener == nil {
		d.openListener = openUsbmuxListener
	}
	d.USB = d.USB.withDefaults()
	return d
}

func (u USBOps) withDefaults() USBOps {
	if u.ListDevices == nil {
		u.ListDevices = ios.ListDevices
	}
	if u.GetValues == nil {
		u.GetValues = ios.GetValues
	}
	if u.GetDevice == nil {
		u.GetDevice = ios.GetDevice
	}
	ports := usbPorts{dial: ios.NewUsbMuxConnectionSimple, listen: net.Listen}
	if u.Probe == nil {
		u.Probe = ports.Probe
	}
	if u.Forward == nil {
		u.Forward = ports.Forward
	}
	if u.RunXCUI == nil {
		u.RunXCUI = testmanagerd.RunXCUIWithBundleIds
	}
	if u.CloseXCUI == nil {
		u.CloseXCUI = testmanagerd.CloseXCUITestRunner
	}
	return u
}
