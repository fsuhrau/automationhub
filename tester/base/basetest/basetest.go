// Package basetest has the fakes the runner tests share: an app connection
// that answers the actions of a test run, and the setup of a run.
package basetest

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"gorm.io/gorm"
)

// AppManager is a device manager whose devices run an app that answers the
// actions of a test run. Every action is recorded like FakeDeviceManager does.
type AppManager struct {
	testkit.FakeDeviceManager
	// Tests is what the app answers to TestsGet.
	Tests []appproto.Test
	// TestPassed is the result the app reports for a started test.
	TestPassed bool
	// Screenshot is what the app answers to GetScreenshot.
	Screenshot []byte
}

// SendAction records the action and lets the app answer it.
func (m *AppManager) SendAction(d device.Device, a action.Interface) error {
	if err := m.FakeDeviceManager.SendAction(d, a); err != nil {
		return err
	}
	if resp := m.answer(a); resp != nil {
		for _, h := range d.ActionHandlers() {
			h.OnActionResponse(d, resp)
		}
	}
	return nil
}

func (m *AppManager) answer(a action.Interface) *appproto.Response {
	resp := &appproto.Response{ActionType: a.GetActionType(), Success: true}
	switch a.GetActionType() {
	case appproto.ActionType_GetTests:
		resp.Payload.Tests = &appproto.Tests{Tests: m.Tests}
	case appproto.ActionType_ExecuteTest:
		resp.ActionType = appproto.ActionType_ExecutionResult
		resp.Success = m.TestPassed
	case appproto.ActionType_GetScreenshot:
		resp.Payload.Screenshot = &appproto.Screenshot{Screenshot: m.Screenshot}
	default:
		return nil
	}
	return resp
}

// PNG returns a tiny valid png image.
func PNG() []byte {
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))) // a buffer does not fail
	return buf.Bytes()
}

// Setup is a run on connected fake devices.
type Setup struct {
	DB      *gorm.DB
	Manager *AppManager
	Devices []*testkit.FakeDevice
	Models  []models.Device
	Deps    base.Deps
	Test    *models.Test
}

// NewSetup returns a database, an app manager and n devices whose apps are
// connected. The timeouts are short and the data dir is a temp dir.
func NewSetup(t *testing.T, n int) *Setup {
	t.Helper()
	db := testkit.DB(t)
	project := testkit.Project(db)
	appModel := testkit.App(db, func(m *models.App) { m.ProjectID = project.ID })
	test := testkit.Test(db, func(m *models.Test) { m.AppID = appModel.ID })
	s := &Setup{DB: db, Manager: &AppManager{TestPassed: true}, Test: test}
	for i := 0; i < n; i++ {
		fake := testkit.NewFakeDevice(string(rune('a' + i)))
		fake.State = device.StateBooted
		fake.AppConnected = true
		fake.Installed = true
		model := testkit.Device(db)
		model.Dev = fake
		s.Devices = append(s.Devices, fake)
		s.Models = append(s.Models, *model)
		s.Manager.Add("fake", fake)
	}
	s.Deps = base.Deps{
		DataDir: t.TempDir(), StoreDir: t.TempDir(),
		DB: db, Devices: s.Manager, NodeURL: "node", ProjectID: project.Identifier, AppID: appModel.ID,
		Timeouts: base.Timeouts{Poll: time.Millisecond, Start: 200 * time.Millisecond, Cancel: time.Second},
	}
	return s
}

// finishTimeout is how long a test waits for a run to finish.
const finishTimeout = 10 * time.Second

// WatchFinished starts listening for finished runs; call it before the run
// starts. The returned func waits for the event of the run.
func WatchFinished(t testing.TB) func(runID uint) events.TestRunFinishedPayload {
	t.Helper()
	got := make(recorder, 64)
	events.TestRunFinished.Register(got)
	return func(runID uint) events.TestRunFinishedPayload {
		t.Helper()
		return watcher{got: got, timeout: finishTimeout}.wait(t, runID)
	}
}

// watcher waits for the finished events of a recorder.
type watcher struct {
	got     <-chan events.TestRunFinishedPayload
	timeout time.Duration
}

func (w watcher) wait(t testing.TB, runID uint) events.TestRunFinishedPayload {
	t.Helper()
	deadline := time.After(w.timeout)
	for {
		select {
		case p := <-w.got:
			if p.TestRunID == runID {
				return p
			}
		case <-deadline:
			t.Fatalf("run %d did not finish", runID)
		}
	}
}

// recorder keeps the finished events; it drops them when nobody reads.
type recorder chan events.TestRunFinishedPayload

func (r recorder) Handle(p events.TestRunFinishedPayload) {
	select {
	case r <- p:
	default:
	}
}
