package base_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/tester/base/basetest"
)

// A cancel stops the running test, waits until the devices are released and
// the run reports as failed-or-finished; the devices are held while it runs.
func TestCancelWaitsUntilTheDevicesAreReleased(t *testing.T) {
	f := newRunner(t, 1, false, nil)
	f.strategy.list = []models.UnityTestFunction{{Method: "m"}}
	running := make(chan struct{})
	f.strategy.onRunTest = func(base.DeviceMap) { close(running) }
	f.strategy.blockUntilCancelled = true
	wait := basetest.WatchFinished(t)

	run, err := f.runner.Run(f.Models, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	<-running
	if got := f.runner.LockedDevices(); len(got) != 1 || got[0] != f.Models[0].ID {
		t.Fatalf("locked devices %v", got)
	}
	if err := f.runner.Cancel("run"); err != nil {
		t.Fatal(err)
	}
	if f.Devices[0].IsLocked() || len(f.runner.LockedDevices()) != 0 {
		t.Fatal("cancel returned before the device was released")
	}
	wait(run.ID)
}

// A cancel that waits too long says so in the run log.
func TestCancelLogsWhenTheRunDoesNotStop(t *testing.T) {
	f := newRunner(t, 1, false, nil)
	f.Deps.Timeouts.Cancel = 10 * time.Millisecond
	f.runner = base.NewRunner(f.Deps, models.TestTypeUnity, func(*base.Runner) base.Strategy { return f.strategy })
	_ = f.runner.Initialize(*f.Test, nil)
	f.strategy.list = []models.UnityTestFunction{{Method: "m"}}
	release := make(chan struct{})
	running := make(chan struct{})
	f.strategy.onRunTest = func(base.DeviceMap) { close(running); <-release }
	wait := basetest.WatchFinished(t)

	run, err := f.runner.Run(f.Models, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	f.run = run
	<-running
	_ = f.runner.Cancel("run")
	close(release)
	wait(run.ID)

	if !f.hasLog("run still stopping") {
		t.Fatal("slow cancel not logged")
	}
}

func TestCancelBeforeTheRunBeganDoesNothing(t *testing.T) {
	f := newRunner(t, 1, false, nil)
	if err := f.runner.Cancel("run"); err != nil {
		t.Fatal(err)
	}
}

type failingStore struct{ base.Store }

func (failingStore) CreateRun(context.Context, *models.TestRun) error { return errors.New("db down") }

func TestRunFailsWhenTheRunCannotBeCreated(t *testing.T) {
	s := basetest.NewSetup(t, 1)
	s.Deps.Store = failingStore{}
	r := base.NewRunner(s.Deps, models.TestTypeUnity, func(*base.Runner) base.Strategy { return &fakeStrategy{} })

	run, err := r.Run(s.Models, nil, "")
	if err == nil || run != nil {
		t.Fatalf("run %v, error %v", run, err)
	}
	if len(s.Devices[0].Calls()) != 0 {
		t.Fatal("devices were used")
	}
}

func TestLogBeforeTheRunBegan(t *testing.T) {
	f := newRunner(t, 1, false, nil)
	f.runner.LogInfo("hello %d", 1)
	var n int64
	f.DB.Model(&models.TestRunLogEntry{}).Where("log = ?", "hello 1").Count(&n)
	if n != 1 {
		t.Fatalf("got %d entries", n)
	}
}

// The unity config names the tests: the app lists them or the config stores them.
func TestListTests(t *testing.T) {
	appTests := []appproto.Test{
		{Class: "A", Method: "a", Categories: []string{"fast"}},
		{Class: "B", Method: "b", Categories: []string{"slow", "net"}},
		{Class: "C", Method: "c"},
	}
	tests := []struct {
		name    string
		unity   *models.TestConfigUnity
		appErr  error
		want    []string
		wantErr bool
	}{
		{"all tests of the app", &models.TestConfigUnity{UnityTestCategoryType: models.AllTest}, nil, []string{"a", "b", "c"}, false},
		{"all of a category", &models.TestConfigUnity{UnityTestCategoryType: models.AllOfCategory, Categories: "net,fast"}, nil, []string{"a", "b"}, false},
		{"selected tests of the config", &models.TestConfigUnity{Model: models.Model{ID: 5}, UnityTestCategoryType: models.SelectedTestsOnly}, nil, []string{"stored"}, false},
		{"app does not answer", &models.TestConfigUnity{UnityTestCategoryType: models.AllTest}, errors.New("gone"), nil, true},
		{"no unity config", nil, nil, nil, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newRunner(t, 1, false, func(c *models.TestConfig) { c.Unity = tc.unity })
			f.DB.Create(&models.UnityTestFunction{TestConfigUnityID: 5, Method: "stored"})
			f.Manager.Tests = appTests
			f.Manager.FailWith("SendAction", tc.appErr)

			got, err := f.runner.ListTests(context.Background(), []base.DeviceMap{{Device: f.Devices[0]}})

			if (err != nil) != tc.wantErr {
				t.Fatalf("error %v, want error %v", err, tc.wantErr)
			}
			if names := methodNames(got); !slices.Equal(names, tc.want) {
				t.Fatalf("got %v, want %v", names, tc.want)
			}
		})
	}
}

func methodNames(tests []models.UnityTestFunction) []string {
	var names []string
	for _, tst := range tests {
		names = append(names, tst.Method)
	}
	return names
}

func TestListTestsWithoutUnityConfigIsATypedError(t *testing.T) {
	f := newRunner(t, 1, false, nil)
	_, err := f.runner.ListTests(context.Background(), nil)
	if !errors.Is(err, base.ErrNoUnityConfig) {
		t.Fatalf("got %v", err)
	}
}

// A cancel while the apps start stops them and fails the run.
func TestCancelWhileTheAppsStart(t *testing.T) {
	f := newRunner(t, 1, true, nil)
	f.Deps.Timeouts.Start = time.Minute
	f.runner = base.NewRunner(f.Deps, models.TestTypeUnity, func(*base.Runner) base.Strategy { return hookedStrategy{f.strategy} })
	_ = f.runner.Initialize(*f.Test, nil)
	f.Devices[0].AppConnected = false
	started := make(chan struct{})
	f.strategy.onAppStarted = func() { close(started) }
	wait := basetest.WatchFinished(t)

	run, err := f.runner.Run(f.Models, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	f.run = run
	<-started
	_ = f.runner.Cancel("run")
	p := wait(run.ID)

	if p.Success || !f.hasLog("no device connected") || f.Devices[0].IsLocked() {
		t.Fatalf("success %v, locked %v", p.Success, f.Devices[0].IsLocked())
	}
	if len(f.Devices[0].CallsTo("StopApp")) < 2 {
		t.Fatal("the starting app was not stopped")
	}
}

// Apps that can't be stopped are logged, before the run and after it.
func TestStopFailuresAreLogged(t *testing.T) {
	f := newRunner(t, 1, false, nil)
	f.Devices[0].FailWith("StopApp", errors.New("adb down"))
	f.start(t, nil)

	logs := f.errorLogs()
	if len(logs) != 2 || logs[0] != "unable to stop app: adb down" || logs[1] != "Unable to stop app: adb down" {
		t.Fatalf("error logs %q", logs)
	}
	if !f.finished.Success {
		t.Fatal("a stop failure must not fail the run")
	}
}
