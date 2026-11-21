package wiring_test

import (
	"io"
	"net"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// testConfig is a service config whose folders live in the test's temp dir.
func testConfig(t *testing.T) config.Service {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return config.Service{
		HostIP: "127.0.0.1",
		Paths:  config.Paths{AppStore: t.TempDir(), Bundles: t.TempDir(), TestData: t.TempDir()},
	}
}

func quietLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logger
}

// busyPort listens on a free port for the whole test, so a server asked to
// listen there fails at once.
func busyPort(t *testing.T) int32 {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return int32(l.Addr().(*net.TCPAddr).Port)
}

// fakeHandler is a device handler of one manager with fixed devices.
type fakeHandler struct {
	name    string
	devices []device.Device
}

func (h *fakeHandler) Name() string                                              { return h.name }
func (h *fakeHandler) Init(string, string, *string) error                        { return nil }
func (h *fakeHandler) Start() error                                              { return nil }
func (h *fakeHandler) Stop() error                                               { return nil }
func (h *fakeHandler) StartDevice(string) error                                  { return nil }
func (h *fakeHandler) StopDevice(string) error                                   { return nil }
func (h *fakeHandler) GetDevices() ([]device.Device, error)                      { return h.devices, nil }
func (h *fakeHandler) RefreshDevices(bool) error                                 { return nil }
func (h *fakeHandler) HasDevice(device.Device) bool                              { return false }
func (h *fakeHandler) RegisterDevice(device.RegisterData) (device.Device, error) { return nil, nil }

func phoneHandler() *fakeHandler {
	return &fakeHandler{name: "android_device", devices: []device.Device{testkit.NewFakeDevice("dev-1")}}
}
