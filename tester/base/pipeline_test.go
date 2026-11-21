package base_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/tester/base/basetest"
)

// fakeStrategy lists the tests it is given and records what runs.
type fakeStrategy struct {
	list      []models.UnityTestFunction
	listErr   error
	panicIn   string
	onRunTest func(dev base.DeviceMap)
	// blockUntilCancelled makes a test wait for the end of the run
	blockUntilCancelled bool

	mu       sync.Mutex
	listed   int
	ran      []action.TestStart
	prepared int
	observed []string
	// onAppStarted is called when an app is started
	onAppStarted func()
}

func (s *fakeStrategy) TestList(context.Context, []base.DeviceMap) ([]models.UnityTestFunction, error) {
	s.mu.Lock()
	s.listed++
	s.mu.Unlock()
	if s.panicIn == "list" {
		panic("list boom")
	}
	return s.list, s.listErr
}

func (s *fakeStrategy) RunTest(ctx context.Context, dev base.DeviceMap, task action.TestStart) {
	if s.panicIn == "run" {
		panic("run boom")
	}
	if s.onRunTest != nil {
		s.onRunTest(dev)
	}
	if s.blockUntilCancelled {
		<-ctx.Done()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ran = append(s.ran, task)
}

// hookedStrategy also prepares the devices and watches the app starts.
type hookedStrategy struct{ *fakeStrategy }

func (s hookedStrategy) Prepare(context.Context, app.Parameter, []base.DeviceMap) {
	s.mu.Lock()
	s.prepared++
	s.mu.Unlock()
}

func (s hookedStrategy) AppStarted(_ context.Context, d device.Device) {
	s.mu.Lock()
	s.observed = append(s.observed, d.DeviceID())
	s.mu.Unlock()
	if s.onAppStarted != nil {
		s.onAppStarted()
	}
}

type fixture struct {
	*basetest.Setup
	runner   *base.Runner
	strategy *fakeStrategy
	run      *models.TestRun
	finished events.TestRunFinishedPayload
}

// newRunner builds a runner of the fake strategy; hooked adds the optional
// strategy interfaces.
func newRunner(t *testing.T, devices int, hooked bool, configure func(*models.TestConfig)) *fixture {
	t.Helper()
	f := &fixture{Setup: basetest.NewSetup(t, devices), strategy: &fakeStrategy{}}
	f.runner = base.NewRunner(f.Deps, models.TestTypeUnity, func(*base.Runner) base.Strategy {
		if hooked {
			return hookedStrategy{f.strategy}
		}
		return f.strategy
	})
	f.Test.TestConfig = models.TestConfig{Type: models.TestTypeUnity}
	if configure != nil {
		configure(&f.Test.TestConfig)
	}
	if err := f.runner.Initialize(*f.Test, map[string]string{"B": "2", "A": "1"}); err != nil {
		t.Fatal(err)
	}
	return f
}

// start runs the test on the devices and waits for the finished event.
func (f *fixture) start(t *testing.T, binary *models.AppBinary) {
	t.Helper()
	wait := basetest.WatchFinished(t)
	run, err := f.runner.Run(f.Models, binary, "")
	if err != nil {
		t.Fatal(err)
	}
	f.run = run
	f.finished = wait(run.ID)
	if !f.runner.WaitFinished(1e9) {
		t.Fatal("run not marked as finished")
	}
}

func (f *fixture) errorLogs() []string {
	var logs []models.TestRunLogEntry
	f.DB.Where("test_run_id = ? AND level = ?", f.run.ID, "error").Find(&logs)
	var out []string
	for _, l := range logs {
		out = append(out, l.Log)
	}
	return out
}

func (f *fixture) hasLog(text string) bool {
	var n int64
	f.DB.Model(&models.TestRunLogEntry{}).Where("test_run_id = ? AND log LIKE ?", f.run.ID, "%"+text+"%").Count(&n)
	return n > 0
}

func methods(d *testkit.FakeDevice, names ...string) []string {
	var out []string
	for _, c := range d.Calls() {
		if slices.Contains(names, c.Method) {
			out = append(out, c.Method)
		}
	}
	return out
}

func bundle() *models.AppBinary {
	return &models.AppBinary{Identifier: "com.example", Name: "app", AppPath: "x.apk"}
}

// The steps run in order: lock, stop, install, start, tests, stop and clear,
// release.
func TestPipelineRunsTheStepsInOrder(t *testing.T) {
	f := newRunner(t, 1, true, nil)
	f.Devices[0].Installed = false
	f.strategy.list = []models.UnityTestFunction{{Class: "A", Method: "m"}}
	f.strategy.onRunTest = func(base.DeviceMap) {
		got := methods(f.Devices[0], "Lock", "StopApp", "InstallApp", "StartApp", "Unlock")
		want := []string{"Lock", "StopApp", "InstallApp", "StartApp"}
		if !slices.Equal(got, want) {
			t.Errorf("before the test: %v, want %v", got, want)
		}
	}
	f.start(t, bundle())

	got := methods(f.Devices[0], "Lock", "StopApp", "ClearAppData", "IsAppInstalled", "InstallApp", "StartApp", "Unlock")
	want := []string{"Lock", "StopApp", "ClearAppData", "IsAppInstalled", "InstallApp", "StartApp", "StopApp", "ClearAppData", "Unlock"}
	if !slices.Equal(got, want) {
		t.Fatalf("calls %v, want %v", got, want)
	}
	if !f.finished.Success || len(f.strategy.ran) != 1 || f.strategy.prepared != 1 || !slices.Equal(f.strategy.observed, []string{"a"}) {
		t.Fatalf("success %v, ran %d, prepared %d, observed %v", f.finished.Success, len(f.strategy.ran), f.strategy.prepared, f.strategy.observed)
	}
	if f.Devices[0].IsLocked() || len(f.runner.LockedDevices()) != 0 {
		t.Fatal("device still locked")
	}
}

// The start of the app, the parameters and the environment reach the device.
func TestPipelineStartsTheAppWithSessionAndEnv(t *testing.T) {
	f := newRunner(t, 1, false, nil)
	f.start(t, bundle())

	calls := f.Devices[0].CallsTo("StartApp")
	if len(calls) != 1 || calls[0].Args[2] != f.run.SessionID || calls[0].Args[3] != "node" {
		t.Fatalf("start app calls %+v", calls)
	}
	if f.run.Parameter != "A=1\nB=2" || f.run.TestID != f.Test.ID {
		t.Fatalf("run %+v", f.run)
	}
}

// A web start without a bundle has nothing to install.
func TestPipelineSkipsTheInstallWithoutABundle(t *testing.T) {
	f := newRunner(t, 1, true, nil)
	f.start(t, nil)

	if got := methods(f.Devices[0], "IsAppInstalled", "InstallApp"); len(got) != 0 || f.strategy.prepared != 0 {
		t.Fatalf("installed: %v, prepared %d", got, f.strategy.prepared)
	}
	if !f.finished.Success {
		t.Fatal("run failed")
	}
}

func TestPipelineFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(f *fixture)
		wantLog   string
		wantCalls []string // device calls among Lock, StartApp, StopApp, Unlock
		wantList  int
	}{
		{"no device available", func(f *fixture) {
			f.Models[0].Dev = nil
		}, "no device available", nil, 0},
		{"lock fails", func(f *fixture) {
			f.Devices[0].FailWith("Lock", errors.New("busy"))
		}, "locking device a failed", []string{"Lock"}, 0},
		{"app does not start", func(f *fixture) {
			f.Devices[0].FailWith("StartApp", errors.New("no space"))
		}, "no device connected", []string{"Lock", "StopApp", "StartApp", "Unlock"}, 0},
		{"app does not connect in time", func(f *fixture) {
			f.Devices[0].AppConnected = false
		}, "Timeout while stating app", []string{"Lock", "StopApp", "StartApp", "StopApp", "Unlock"}, 0},
		{"test list fails", func(f *fixture) {
			f.strategy.listErr = errors.New("no list")
		}, "Get tests failed: no list", []string{"Lock", "StopApp", "StartApp", "StopApp", "Unlock"}, 1},
		{"test list panics", func(f *fixture) {
			f.strategy.panicIn = "list"
		}, "recovered: list boom", []string{"Lock", "StopApp", "StartApp", "StopApp", "Unlock"}, 1},
		{"test panics", func(f *fixture) {
			f.strategy.list = []models.UnityTestFunction{{Method: "m"}}
			f.strategy.panicIn = "run"
		}, "recovered: run boom", []string{"Lock", "StopApp", "StartApp", "StopApp", "Unlock"}, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newRunner(t, 1, false, nil)
			tc.prepare(f)
			f.start(t, nil)

			if f.finished.Success {
				t.Fatal("run reported as success")
			}
			if logs := f.errorLogs(); !slices.ContainsFunc(logs, func(l string) bool { return strings.Contains(l, tc.wantLog) }) {
				t.Fatalf("error logs %q lack %q", logs, tc.wantLog)
			}
			if got := methods(f.Devices[0], "Lock", "StartApp", "StopApp", "Unlock"); !slices.Equal(got, tc.wantCalls) {
				t.Fatalf("calls %v, want %v", got, tc.wantCalls)
			}
			if f.strategy.listed != tc.wantList {
				t.Fatalf("listed %d times, want %d", f.strategy.listed, tc.wantList)
			}
		})
	}
}

// A device that is not connected is skipped, the others run.
func TestPipelineSkipsUnavailableDevices(t *testing.T) {
	f := newRunner(t, 2, false, nil)
	f.Devices[1].State = device.StateUnauthorized
	f.strategy.list = []models.UnityTestFunction{{Method: "m"}}
	f.start(t, nil)

	if !f.finished.Success || len(f.strategy.ran) != 1 {
		t.Fatalf("success %v, ran %d", f.finished.Success, len(f.strategy.ran))
	}
	if len(f.Devices[1].CallsTo("Lock")) != 0 || !f.hasLog("skipping device") {
		t.Fatal("unavailable device was not skipped")
	}
}

func TestPipelineStartsShutDownDevices(t *testing.T) {
	f := newRunner(t, 1, false, nil)
	f.Devices[0].State = device.StateShutdown
	f.Manager.FailWith("Start", errors.New("no emulator"))
	f.start(t, nil)

	if len(f.Manager.CallsTo("Start")) != 1 || !f.hasLog("unable to start device") {
		t.Fatalf("manager calls %+v", f.Manager.Calls())
	}
	if len(f.Manager.CallsTo("Stop")) != 1 {
		t.Fatal("device not stopped on release")
	}
}

// flakyDevice fails its first installations.
type flakyDevice struct {
	*testkit.FakeDevice
	failures int
}

func (d *flakyDevice) InstallApp(p *app.Parameter) error {
	err := d.FakeDevice.InstallApp(p)
	if d.failures > 0 {
		d.failures--
		return errors.New("conflict")
	}
	return err
}

func TestPipelineInstallRetriesAfterUninstall(t *testing.T) {
	tests := []struct {
		name          string
		failures      int
		wantUninstall int
		wantLog       string
	}{
		{"second attempt works", 1, 1, ""},
		{"both attempts fail", 2, 2, "unable to install app"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newRunner(t, 1, false, nil)
			f.Devices[0].Installed = false
			f.Models[0].Dev = &flakyDevice{FakeDevice: f.Devices[0], failures: tc.failures}
			f.start(t, bundle())

			if got := len(f.Devices[0].CallsTo("InstallApp")); got != 2 {
				t.Fatalf("%d install attempts, want 2", got)
			}
			if got := len(f.Devices[0].CallsTo("UninstallApp")); got != tc.wantUninstall {
				t.Fatalf("%d uninstalls, want %d", got, tc.wantUninstall)
			}
			if tc.wantLog != "" && !f.hasLog(tc.wantLog) {
				t.Fatalf("missing log %q", tc.wantLog)
			}
		})
	}
}

func TestPipelineLogsAFailedInstallationCheck(t *testing.T) {
	f := newRunner(t, 1, false, nil)
	f.Devices[0].FailWith("IsAppInstalled", errors.New("adb down"))
	f.start(t, bundle())

	if len(f.Devices[0].CallsTo("InstallApp")) != 0 || !f.hasLog("check installation failed") {
		t.Fatal("installation check error not handled")
	}
}

// With recording on, every device is recorded until the run ends.
func TestPipelineRecordsTheSessionWhenAsked(t *testing.T) {
	f := newRunner(t, 2, false, func(c *models.TestConfig) { c.RecordSession = true })
	f.start(t, nil)

	for _, d := range f.Devices {
		if len(d.CallsTo("StartVideo")) == 0 {
			t.Fatalf("device %s was not recorded", d.DeviceID())
		}
	}
	if !f.finished.Success {
		t.Fatal("run failed")
	}
}

func TestInitializeRejectsAnotherType(t *testing.T) {
	f := newRunner(t, 1, false, nil)
	other := models.Test{TestConfig: models.TestConfig{Type: models.TestTypeScenario}}
	if err := f.runner.Initialize(other, nil); err == nil {
		t.Fatal("type mismatch accepted")
	}
}
