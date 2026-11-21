package run_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

func unityTest() *models.Test {
	return &models.Test{
		Model: models.Model{ID: 3}, AppID: 2, App: &models.App{Platform: models.PlatformTypeAndroid},
		TestConfig: models.TestConfig{
			Model: models.Model{ID: 8}, Type: models.TestTypeUnity, AllDevices: true, Unity: &models.TestConfigUnity{},
		},
	}
}

func storedDevice(id uint, identifier string) models.Device {
	return models.Device{Model: models.Model{ID: id}, DeviceIdentifier: identifier, Name: "Pixel " + identifier}
}

// startEnv is an env with a startable unity test, one bundle and one live
// device.
func startEnv() *env {
	e := newEnv()
	e.repo.test = unityTest()
	e.repo.bundle = &models.AppBinary{Model: models.Model{ID: 4}}
	e.repo.devices = []models.Device{storedDevice(1, "dev-1")}
	e.pool.Add("android_device", testkit.NewFakeDevice("dev-1"))
	return e
}

func request(e *env) run.StartRequest {
	return run.StartRequest{Project: e.project, TestID: 3, BundleID: 4, StartURL: "https://x.test", Params: "a=1;b=2"}
}

type startedRecorder chan events.TestRunStartedPayload

func (r startedRecorder) Handle(p events.TestRunStartedPayload) {
	select {
	case r <- p:
	default:
	}
}

func TestStartRunsTheTestOnTheDevicesOfItsConfiguration(t *testing.T) {
	e := startEnv()
	seen := make(startedRecorder, 8)
	events.TestRunStarted.Register(seen)

	ref, err := e.svc.Start(bg(), request(e))

	if err != nil || ref.Run.ID != 5 {
		t.Fatalf("ref %+v, err %v", ref, err)
	}
	r := e.runner
	if r.env["a"] != "1" || r.env["b"] != "2" || r.test.ID != 3 || r.startURL != "https://x.test" || r.binary.ID != 4 {
		t.Errorf("runner got env %v test %d url %q binary %+v", r.env, r.test.ID, r.startURL, r.binary)
	}
	if len(r.devices) != 1 || r.devices[0].Dev == nil {
		t.Errorf("the run got devices %+v without their live device", r.devices)
	}
	if e.kinds[0] != models.TestTypeUnity || !e.repo.deviceQ.All {
		t.Errorf("runner kinds %v, device query %+v", e.kinds, e.repo.deviceQ)
	}
	if _, ok := e.runners[5]; !ok || len(e.tracked) != 1 || e.tracked[0] != 5 {
		t.Errorf("registered %v, tracked %v", e.runners, e.tracked)
	}
	select {
	case p := <-seen:
		if p.ProjectID != "proj" || p.AppID != 2 || p.TestID != 3 || p.TestRunID != 5 {
			t.Errorf("event %+v", p)
		}
	case <-time.After(time.Second):
		t.Fatal("no TestRunStarted event")
	}
}

func TestStartOnTheDevicesOfAnExplicitList(t *testing.T) {
	e := startEnv()
	e.repo.test.TestConfig.AllDevices = false
	e.repo.test.TestConfig.Devices = []models.TestConfigDevice{{DeviceID: 1}, {DeviceID: 9}}

	if _, err := e.svc.Start(bg(), request(e)); err != nil {
		t.Fatal(err)
	}

	if q := e.repo.deviceQ; q.All || len(q.IDs) != 2 || q.IDs[1] != 9 || q.Project != e.project {
		t.Fatalf("device query %+v", q)
	}
}

func TestStartRunsWebAndEditorAppsWithoutABundle(t *testing.T) {
	for _, platform := range []models.PlatformType{models.PlatformTypeWeb, models.PlatformTypeEditor} {
		e := startEnv()
		e.repo.test.App.Platform = platform
		e.repo.fail["Bundle"] = apperr.NotFound("must not be asked")

		if _, err := e.svc.Start(bg(), request(e)); err != nil {
			t.Fatalf("platform %v: %v", platform, err)
		}

		if e.runner.binary != nil {
			t.Errorf("platform %v: binary %+v", platform, e.runner.binary)
		}
	}
}

func TestStartRunsScenarioTests(t *testing.T) {
	e := startEnv()
	e.repo.test.TestConfig.Type, e.repo.test.TestConfig.Unity = models.TestTypeScenario, nil

	if _, err := e.svc.Start(bg(), request(e)); err != nil {
		t.Fatal(err)
	}

	if e.kinds[0] != models.TestTypeScenario {
		t.Fatalf("runner kinds %v", e.kinds)
	}
}

func TestStartErrors(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name  string
		setup func(*env)
		want  apperr.Kind
	}{
		{"unknown test", func(e *env) { e.repo.fail["StartTest"] = apperr.NotFound("test") }, apperr.KindNotFound},
		{"unknown bundle", func(e *env) { e.repo.fail["Bundle"] = apperr.NotFound("bundle") }, apperr.KindNotFound},
		{"archived bundle", func(e *env) {
			now := time.Now()
			e.repo.bundle.ArchivedAt = &now
		}, apperr.KindInvalid},
		{"devices load", func(e *env) { e.repo.fail["Devices"] = apperr.Internal("db") }, apperr.KindInternal},
		{"test without devices", func(e *env) { e.repo.devices = nil }, apperr.KindConflict},
		{"no usable device", func(e *env) { e.repo.devices = []models.Device{storedDevice(7, "unknown")} }, apperr.KindConflict},
		{"unity test without settings", func(e *env) { e.repo.test.TestConfig.Unity = nil }, apperr.KindInternal},
		{"unknown test type", func(e *env) { e.newErr = apperr.Internal("invalid test config type") }, apperr.KindInternal},
		{"initialize fails", func(e *env) { e.runner.failOn, e.runner.err = "Initialize", boom }, apperr.KindInternal},
		{"run fails", func(e *env) { e.runner.failOn, e.runner.err = "Run", boom }, apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := startEnv()
			tc.setup(e)

			ref, err := e.svc.Start(bg(), request(e))

			assertKind(t, err, tc.want)
			if ref.Run != nil || len(e.runners) != 0 || len(e.tracked) != 0 {
				t.Fatalf("a failed start left ref %+v, runners %v, tracked %v", ref, e.runners, e.tracked)
			}
		})
	}
}

func TestStartRefusesARunWhoseDevicesAreAllUnavailable(t *testing.T) {
	e := startEnv()
	gone := testkit.NewFakeDevice("dev-2")
	gone.State = device.StateRemoteDisconnected
	e.pool.Add("android_device", gone)
	e.repo.devices = []models.Device{storedDevice(2, "dev-2"), {Model: models.Model{ID: 3}, DeviceIdentifier: "emulator-5554"}}

	_, err := e.svc.Start(bg(), request(e))

	assertKind(t, err, apperr.KindConflict)
	for _, want := range []string{"Pixel dev-2 is disconnected", "emulator-5554 is not connected to the hub", "no device available to run the test"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%q missing in %q", want, err)
		}
	}
}

func TestStartSkipsUnavailableDevicesNextToAvailableOnes(t *testing.T) {
	e := startEnv()
	e.repo.devices = append(e.repo.devices, models.Device{Model: models.Model{ID: 3}, DeviceIdentifier: "emulator-5554"})

	if _, err := e.svc.Start(bg(), request(e)); err != nil {
		t.Fatal(err)
	}

	if len(e.runner.devices) != 2 {
		t.Fatalf("the run got %d devices", len(e.runner.devices))
	}
}
