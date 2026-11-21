package mcp

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

// RunService starts, cancels and reports test runs of a project.
type RunService interface {
	Start(ctx context.Context, req runsvc.StartRequest) (runsvc.RunRef, error)
	Cancel(ctx context.Context, projectID, runID uint) error
	Report(ctx context.Context, projectID, runID uint) (*runsvc.Report, error)
	Progress(ctx context.Context, projectID, runID uint) (runsvc.Progress, error)
}

// ProtocolService reads the protocols of a project and their logs.
type ProtocolService interface {
	ProtocolDetail(ctx context.Context, projectID, protocolID uint) (*runsvc.ProtocolDetail, error)
	ProtocolData(ctx context.Context, projectID, protocolID uint) ([]models.ProtocolEntry, error)
	Logs(ctx context.Context, q runsvc.LogQuery) (*runsvc.LogPage, error)
}

// DeviceLookup finds the devices a project can see.
type DeviceLookup interface {
	Devices(ctx context.Context, project *models.Project) ([]models.Device, error)
	Device(ctx context.Context, project *models.Project, id uint) (*models.Device, error)
}

// AppService reads the apps and bundles of a project and stores new bundles.
type AppService interface {
	App(ctx context.Context, sc appsvc.Scope) (*models.App, error)
	Apps(ctx context.Context, projectID uint) ([]models.App, error)
	Bundles(ctx context.Context, sc appsvc.Scope) ([]appsvc.BundleEntry, error)
	StoreBundle(ctx context.Context, sc appsvc.Scope, filePath string) (*models.AppBinary, error)
}

// TestService lists the tests of an app.
type TestService interface {
	List(ctx context.Context, sc testsvc.Scope) ([]testsvc.ListEntry, error)
}

// DeviceService is the device use cases the tools share with the REST api.
type DeviceService interface {
	ControlApp(ctx context.Context, project *models.Project, binaryID uint) (*app.Parameter, error)
	Unlock(ctx context.Context, project *models.Project, id string) (*models.Device, error)
}

// Leases hands out the exclusive control of devices.
type Leases interface {
	Acquire(dev device.Device, owner string) (hubcontrol.Lease, error)
	Release(deviceID, token string) error
	TokenFor(deviceID, owner string) (string, error)
	Holder(deviceID string) (hubcontrol.Lease, bool)
	ReleaseOwner(owner string) int
}

// Operations are what can be done to a device that is controlled live; the
// ones that change it need the token of the lease.
type Operations interface {
	Screenshot(dev device.Device) ([]byte, int, int, error)
	AppScreenshot(ctx context.Context, dev device.Device) ([]byte, int, int, error)
	Hierarchy(ctx context.Context, dev device.Device, pretty bool) ([]byte, error)
	Logs(dev device.Device, lines int) ([]string, error)
	Tap(dev device.Device, token string, x, y int) error
	Swipe(dev device.Device, token string, fromX, fromY, toX, toY int, duration time.Duration) error
	InputText(dev device.Device, token, text string) error
	PressKey(dev device.Device, token, key string) error
	TouchElement(ctx context.Context, dev device.Device, token string, ref hubcontrol.ElementRef) error
	InstallApp(ctx context.Context, dev device.Device, token string, params *app.Parameter) error
	StartApp(dev device.Device, token string, params *app.Parameter, sessionID, nodeURL string) error
	StopApp(dev device.Device, token string, params *app.Parameter, clearData bool) error
}

// Control is the live control of devices: the leases and the operations.
// *hubcontrol.Service implements it.
type Control interface {
	Leases
	Operations
}

// Pool is the pool of live devices.
type Pool interface {
	GetDevice(id string) (device.Device, string)
}
