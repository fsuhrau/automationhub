package base

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/protocol"
)

// DeviceMap is a device of the run with its database model.
type DeviceMap struct {
	Device device.Device
	Model  models.Device
}

// Runner runs one test through the pipeline; the strategy decides what
// tests there are and how one is run. It implements tester.Interface. A
// runner serves one run.
type Runner struct {
	Deps
	// Env, Config and Test are set by Initialize.
	Env    map[string]string
	Config models.TestConfig
	Test   models.Test

	// ProtocolWriter and TestRun exist once the run is started.
	ProtocolWriter *protocol.ProtocolWriter
	TestRun        models.TestRun

	kind     models.TestType
	strategy Strategy

	ctlMu  sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	startupMu sync.Mutex

	lockedMu sync.Mutex
	locked   []uint

	// closed once the run stopped its apps and released its devices
	finished     chan struct{}
	finishedOnce sync.Once
}

// NewRunner builds a runner for tests of the type; newStrategy gets the
// runner so the strategy can log and reach its dependencies.
func NewRunner(deps Deps, kind models.TestType, newStrategy func(*Runner) Strategy) *Runner {
	r := &Runner{Deps: deps.withDefaults(), kind: kind, finished: make(chan struct{})}
	r.strategy = newStrategy(r)
	return r
}

// Initialize sets the test and the environment of the run.
func (r *Runner) Initialize(test models.Test, env map[string]string) error {
	if test.TestConfig.Type != r.kind {
		return fmt.Errorf("test config type %d does not match the runner type %d", test.TestConfig.Type, r.kind)
	}
	r.Env = env
	r.Test = test
	r.Config = test.TestConfig
	return nil
}

// Run creates the run and executes it in the background.
func (r *Runner) Run(devs []models.Device, binary *models.AppBinary, startURL string) (*models.TestRun, error) {
	ctx := r.begin()
	var binaryID uint
	if binary != nil {
		binaryID = binary.ID
	}
	if err := r.InitNewTestSession(ctx, Session{AppBinaryID: binaryID, StartURL: startURL, Parameter: r.envLines()}); err != nil {
		r.stopContext()
		return nil, err
	}
	go r.Execute(ctx, runInput{devs: devs, binary: binary, startURL: startURL})
	return &r.TestRun, nil
}

// begin opens the context of the run; Cancel ends it.
func (r *Runner) begin() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	r.ctlMu.Lock()
	defer r.ctlMu.Unlock()
	r.ctx, r.cancel = ctx, cancel
	return ctx
}

// envLines is the run's environment as "key=value" lines.
func (r *Runner) envLines() string {
	lines := make([]string, 0, len(r.Env))
	for _, k := range slices.Sorted(maps.Keys(r.Env)) {
		lines = append(lines, fmt.Sprintf("%s=%s", k, r.Env[k]))
	}
	return strings.Join(lines, "\n")
}

// stopContext ends the context of the run; it reports whether a run was begun.
func (r *Runner) stopContext() bool {
	r.ctlMu.Lock()
	cancel := r.cancel
	r.ctlMu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}

// Cancel stops the run and waits until it released its devices, so the test
// can be started again right away.
func (r *Runner) Cancel(string) error {
	if r.stopContext() && !r.WaitFinished(r.Timeouts.Cancel) {
		r.LogError("run still stopping %s after the cancel", r.Timeouts.Cancel)
	}
	return nil
}

// Finished marks the run as done with its devices; call it when the run's
// work ends, however it ends.
func (r *Runner) Finished() {
	r.finishedOnce.Do(func() { close(r.finished) })
}

// WaitFinished waits until the run released its devices, at most timeout.
func (r *Runner) WaitFinished(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-r.finished:
		return true
	case <-timer.C:
		return false
	}
}

// scheduler hands the tests to the strategy.
func (r *Runner) scheduler() Scheduler {
	return Scheduler{Mode: r.Config.ExecutionType, Env: r.Env, Run: r.strategy.RunTest}
}
