// Package device holds the use cases around the devices of a project: the
// list and detail views with their live state, editing, deleting, unlocking,
// running tests on a device and the results that ran on it. It knows no
// HTTP and no gorm; the storage sits behind Repo.
package device

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Repo is the storage of devices. Every lookup is scoped to the project.
type Repo interface {
	List(ctx context.Context, project *models.Project, platform *models.PlatformType) ([]models.Device, error)
	Find(ctx context.Context, project *models.Project, id string) (*models.Device, error)
	FindDetail(ctx context.Context, project *models.Project, id string) (*models.Device, error)
	Delete(ctx context.Context, d *models.Device) error
	Apply(ctx context.Context, deviceID uint, change Change) error
	Protocols(ctx context.Context, scope ProtocolScope) ([]models.TestProtocol, error)
	Bundle(ctx context.Context, project *models.Project, id uint) (*models.AppBinary, error)
}

// Devices is the pool of live devices.
type Devices interface {
	GetDevice(id string) (device.Device, string)
	SendAction(dev device.Device, a action.Interface) error
}

// Leases tells who controls a device live.
type Leases interface {
	Holder(deviceID string) (control.Lease, bool)
	ForceRelease(deviceID string) bool
}

// Executor runs one test on a device and waits for it to finish.
type Executor interface {
	Execute(ctx context.Context, dev device.Device, test action.TestStart, timeout time.Duration) error
}

// Locker locks and unlocks a device and announces the change.
type Locker interface {
	SetLock(dev device.Device, locked bool) error
}

// busLocker announces lock changes on an event bus.
type busLocker struct {
	bus *events.Bus[events.DeviceLockChangedPayload]
}

func (l busLocker) SetLock(dev device.Device, locked bool) error {
	return events.SetDeviceLock(l.bus, dev, locked)
}

// Deps are what a Service needs. NewExecutor returns a fresh executor per
// call because an executor keeps the state of one run.
type Deps struct {
	Repo        Repo
	Devices     Devices
	Leases      Leases
	NewExecutor func() Executor
	// Locker defaults to the global device lock bus.
	Locker Locker
	// StoreDir is the folder of the uploaded bundles, for the install
	// parameters of an app.
	StoreDir string
}

// Service is the device use cases.
type Service struct {
	deps Deps
}

// New returns a Service on the dependencies.
func New(deps Deps) *Service {
	if deps.Locker == nil {
		deps.Locker = busLocker{bus: events.DeviceLockChanged}
	}
	return &Service{deps: deps}
}
