package base

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/runnerstore"
	"gorm.io/gorm"
)

// Store is the repository of a run: the runners never call gorm directly.
type Store interface {
	CreateRun(ctx context.Context, run *models.TestRun) error
	AddLog(ctx context.Context, entry *models.TestRunLogEntry) error
	UnityTests(ctx context.Context, unityConfigID uint) ([]models.UnityTestFunction, error)
}

// Timeouts bound the waits of a run; zero values use the defaults.
type Timeouts struct {
	// Poll is how often the connection of a started app is checked.
	Poll time.Duration
	// Start is how long all apps get to connect.
	Start time.Duration
	// Cancel is how long a cancel waits for the run to release its devices.
	Cancel time.Duration
}

const (
	defaultPoll   = 500 * time.Millisecond
	defaultStart  = time.Minute
	defaultCancel = 30 * time.Second
)

func (t Timeouts) withDefaults() Timeouts {
	if t.Poll == 0 {
		t.Poll = defaultPoll
	}
	if t.Start == 0 {
		t.Start = defaultStart
	}
	if t.Cancel == 0 {
		t.Cancel = defaultCancel
	}
	return t
}

// Deps are the dependencies of a runner.
type Deps struct {
	// DB is handed to the protocol writer and the recordings; the runner
	// itself reads and writes through Store.
	DB *gorm.DB
	// Store defaults to the gorm store on DB.
	Store   Store
	Devices manager.Devices
	NodeURL string
	// ProjectID and AppID name the project and app of the test.
	ProjectID string
	AppID     uint
	// Clock defaults to the real clock.
	Clock    device.Clock
	Timeouts Timeouts
}

func (d Deps) withDefaults() Deps {
	if d.Store == nil {
		d.Store = runnerstore.New(d.DB)
	}
	if d.Clock == nil {
		d.Clock = device.RealClock{}
	}
	d.Timeouts = d.Timeouts.withDefaults()
	return d
}
