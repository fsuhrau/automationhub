package testkit

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/videostream/videostreamtest"
)

// Compile-time checks: FakeDevice implements the device interface and every
// optional capability of package device.
var (
	_ device.Device         = (*FakeDevice)(nil)
	_ device.AppDataClearer = (*FakeDevice)(nil)
	_ device.InputCapable   = (*FakeDevice)(nil)
	_ device.LogReader      = (*FakeDevice)(nil)
	_ device.LogFollower    = (*FakeDevice)(nil)
	_ device.VideoStreamer  = (*FakeDevice)(nil)
)

// FakeDevice is a device that records every call. The exported fields are the
// values it reports; set them before use. FailWith scripts errors per method
// name, for example FailWith("InstallApp", err).
type FakeDevice struct {
	recorder

	ID, Name, OSName, OSVersion, Target string
	Parameter                           map[string]string
	Type, Platform                      int
	State                               device.State
	Installed, AppConnected             bool
	Timeout                             time.Duration
	Features                            []string
	PNG                                 []byte
	Width, Height                       int
	LogLines                            []string

	locked   bool
	conn     *device.Connection
	writer   device.LogWriter
	handlers []action.ActionHandler
	cancel   context.CancelFunc
}

// NewFakeDevice returns a device with the given id and the name "fake".
func NewFakeDevice(id string) *FakeDevice {
	return &FakeDevice{ID: id, Name: "fake", State: device.StateNone}
}

func (d *FakeDevice) DeviceID() string                   { return d.ID }
func (d *FakeDevice) DeviceState() device.State          { return d.State }
func (d *FakeDevice) DeviceOSName() string               { return d.OSName }
func (d *FakeDevice) DeviceOSVersion() string            { return d.OSVersion }
func (d *FakeDevice) DeviceName() string                 { return d.Name }
func (d *FakeDevice) DeviceParameter() map[string]string { return d.Parameter }
func (d *FakeDevice) DeviceType() int                    { return d.Type }
func (d *FakeDevice) PlatformType() int                  { return d.Platform }
func (d *FakeDevice) TargetVersion() string              { return d.Target }
func (d *FakeDevice) ConnectionTimeout() time.Duration   { return d.Timeout }
func (d *FakeDevice) IsAppConnected() bool               { return d.AppConnected }

func (d *FakeDevice) IsAppInstalled(p *app.Parameter) (bool, error) {
	return d.Installed, d.record("IsAppInstalled", p)
}
func (d *FakeDevice) InstallApp(p *app.Parameter) error   { return d.record("InstallApp", p) }
func (d *FakeDevice) UninstallApp(p *app.Parameter) error { return d.record("UninstallApp", p) }
func (d *FakeDevice) ClearAppData(p *app.Parameter) error { return d.record("ClearAppData", p) }
func (d *FakeDevice) StopApp(p *app.Parameter) error      { return d.record("StopApp", p) }

//nolint:revive // implements device.Device
func (d *FakeDevice) StartApp(cfg *device.DeviceConfig, p *app.Parameter, sessionID, nodeURL string) error {
	return d.record("StartApp", cfg, p, sessionID, nodeURL)
}

// NewContext returns a context that Cancel cancels.
func (d *FakeDevice) NewContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	d.mu.Lock()
	d.cancel = cancel
	d.mu.Unlock()
	return ctx
}

// Cancel cancels the context of the last NewContext.
func (d *FakeDevice) Cancel() {
	d.mu.Lock()
	cancel := d.cancel
	d.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	_ = d.record("Cancel")
}

func (d *FakeDevice) SetConnection(c *device.Connection) {
	d.mu.Lock()
	d.conn = c
	d.mu.Unlock()
	_ = d.record("SetConnection", c)
}

func (d *FakeDevice) Connection() *device.Connection {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.conn
}

func (d *FakeDevice) Send(b []byte) error           { return d.record("Send", b) }
func (d *FakeDevice) StartRecording(p string) error { return d.record("StartRecording", p) }
func (d *FakeDevice) StopRecording() error          { return d.record("StopRecording") }
func (d *FakeDevice) Execute(s string)              { _ = d.record("Execute", s) }
func (d *FakeDevice) RunNativeScript(b []byte)      { _ = d.record("RunNativeScript", b) }

func (d *FakeDevice) GetScreenshot() ([]byte, int, int, error) {
	return d.PNG, d.Width, d.Height, d.record("GetScreenshot")
}

func (d *FakeDevice) HasFeature(f string) bool {
	_ = d.record("HasFeature", f)
	for _, have := range d.Features {
		if have == f {
			return true
		}
	}
	return false
}

func videoStream(w, h int) device.VideoStream {
	return videostreamtest.New(device.VideoMeta{Codec: device.CodecMJPEG, Width: w, Height: h, ScreenWidth: w, ScreenHeight: h})
}
