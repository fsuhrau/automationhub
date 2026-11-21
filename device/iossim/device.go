package iossim

import (
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/storage/models"
)

const ConnectionTimeout = 10 * time.Second

// Deps are the injected dependencies of the handler and its devices. Zero
// fields get their defaults.
type Deps struct {
	Runner device.CommandRunner
	Clock  device.Clock
	// KeepAppRunning makes StartApp and StopApp do nothing, so an app that
	// runs in the simulator already is left alone.
	KeepAppRunning bool
}

func (d Deps) withDefaults() Deps {
	if d.Runner == nil {
		d.Runner = device.ExecRunner{}
	}
	if d.Clock == nil {
		d.Clock = device.RealClock{}
	}
	return d
}

type Device struct {
	generic.Device
	generic.Info
	deps         Deps
	simctl       simctl
	lastUpdateAt time.Time

	recMu     sync.Mutex
	recording device.Process
}

// NewDevice returns a simulator that is shut down until a refresh says otherwise.
func NewDevice(info generic.Info, deps Deps) *Device {
	deps = deps.withDefaults()
	info.Type = int(models.DeviceTypePhone)
	info.Platform = int(models.PlatformTypeiOSSimulator)
	return &Device{Info: info, deps: deps, simctl: simctl{runner: deps.Runner}}
}

// LastUpdate is the time of the refresh that last saw the simulator.
func (d *Device) LastUpdate() time.Time {
	return d.lastUpdateAt
}

func (d *Device) IsAppConnected() bool {
	return d.Connection() != nil
}

func (d *Device) ConnectionTimeout() time.Duration {
	return ConnectionTimeout
}
