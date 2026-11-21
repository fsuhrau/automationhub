package base

import (
	"context"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/storage/models"
)

// runInput is what the caller of a run hands to the pipeline.
type runInput struct {
	devs     []models.Device
	binary   *models.AppBinary
	startURL string
}

// runState is what the steps of a run share.
type runState struct {
	runInput
	params    app.Parameter
	devices   []DeviceMap
	connected []DeviceMap
	cleanups  []func()
}

// onExit registers cleanup work; it runs in reverse order when the run ends,
// however it ends.
func (s *runState) onExit(cleanup func()) {
	s.cleanups = append(s.cleanups, cleanup)
}

func (s *runState) unwind() {
	for i := len(s.cleanups) - 1; i >= 0; i-- {
		s.cleanups[i]()
	}
}

// step does one part of a run; it returns false when the run can't go on.
type step func(ctx context.Context, s *runState) bool

// steps is the pipeline: lock, start, record, install, connect, run. Stopping
// and clearing the app, stopping the recordings and releasing the devices are
// the cleanups the steps register.
func (r *Runner) steps() []step {
	return []step{r.lock, r.startDevices, r.record, r.install, r.connect, r.runTests}
}

// Execute runs the pipeline for the run and finishes it; it never panics.
func (r *Runner) Execute(ctx context.Context, in runInput) {
	defer r.Finished()
	// a run without devices finishes as well; it used to stay running
	defer r.TestSessionFinished()
	defer r.recoverRun()

	s := &runState{runInput: in, params: GetParams(in.binary, in.startURL)}
	defer s.unwind()
	for _, step := range r.steps() {
		if !step(ctx, s) {
			return
		}
	}
}

// recoverRun turns a panic of the run into a failed run; defer it directly.
func (r *Runner) recoverRun() {
	if err := recover(); err != nil {
		r.FailRun("test execution failed, recovered: %v", err)
	}
}

func (r *Runner) lock(_ context.Context, s *runState) bool {
	s.devices = r.LockDevices(s.devs)
	if len(s.devices) == 0 {
		r.FailRun("no device available to run the tests")
		return false
	}
	s.onExit(func() { r.UnlockDevices(s.devices) })
	return true
}

func (r *Runner) startDevices(ctx context.Context, s *runState) bool {
	r.LogInfo("Starting devices")
	r.StartDevices(ctx, s.devices)
	return true
}

// record stops before the devices are unlocked, so the files are complete
// when the run finishes.
func (r *Runner) record(_ context.Context, s *runState) bool {
	s.onExit(r.RecordSessions(s.devices))
	return true
}

func (r *Runner) install(ctx context.Context, s *runState) bool {
	r.LogInfo("Stop apps if running")
	r.stopApps(ctx, s.params, s.devices)
	if s.params.App == nil {
		return true
	}
	if p, ok := r.strategy.(Preparer); ok {
		p.Prepare(ctx, s.params, s.devices)
	}
	r.LogInfo("Install app on devices")
	r.installApps(ctx, s.params, s.devices)
	return true
}

func (r *Runner) connect(ctx context.Context, s *runState) bool {
	r.LogInfo("Start app on devices and wait for connection")
	connected, err := r.startApps(ctx, s)
	s.connected = connected
	if !r.checkStarted(connected, err) {
		return false
	}
	s.onExit(func() { r.stopConnected(s) })
	return true
}

func (r *Runner) stopConnected(s *runState) {
	r.LogInfo("Stop apps")
	for _, d := range s.connected {
		if err := StopAndClearApp(d.Device, &s.params); err != nil {
			r.LogError("Unable to stop app: %v", err)
		}
	}
}

func (r *Runner) runTests(ctx context.Context, s *runState) bool {
	r.LogInfo("Get test list")
	tests, err := r.strategy.TestList(ctx, s.connected)
	if err != nil {
		r.FailRun("Get tests failed: %v", err)
		return false
	}
	if len(tests) == 0 {
		r.LogInfo("No Tests")
	}
	r.LogInfo("Execute Tests")
	r.scheduler().Schedule(ctx, s.connected, tests)
	return true
}
