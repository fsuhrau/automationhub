package runstore_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/service/protocol"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/runstore"
	"gorm.io/gorm"
)

func bg() context.Context { return context.Background() }

// world is two projects with one app, test, run and measuring protocol
// each; the foreign one is created first, so its ids are the lower ones.
type world struct {
	db                     *gorm.DB
	store                  *runstore.Store
	device                 *models.Device
	mineProject, foreProj  *models.Project
	mineApp, foreApp       *models.App
	mine, foreign          *models.Test
	run1, run2, foreignRun *models.TestRun
	proto1, proto2, foreP  *models.TestProtocol
}

func newWorld(t *testing.T) *world {
	t.Helper()
	db := testkit.DB(t)
	w := &world{db: db, store: runstore.New(db), device: testkit.Device(db)}
	w.foreProj = testkit.Project(db, func(p *models.Project) { p.Identifier = "foreign" })
	w.foreApp = testkit.App(db, func(a *models.App) { a.ProjectID = w.foreProj.ID })
	w.foreign = testkit.Test(db, func(x *models.Test) { x.AppID = w.foreApp.ID })
	w.foreignRun = w.run(w.foreign, 1000)
	w.foreP = w.protocol(w.foreignRun, 100)
	w.mineProject = testkit.Project(db, func(p *models.Project) { p.Identifier = "mine" })
	w.mineApp = testkit.App(db, func(a *models.App) { a.ProjectID = w.mineProject.ID })
	w.mine = testkit.Test(db, func(x *models.Test) { x.AppID = w.mineApp.ID })
	w.run1 = w.run(w.mine, 20)
	w.proto1 = w.protocol(w.run1, 10)
	w.run2 = w.run(w.mine, 30)
	w.proto2 = w.protocol(w.run2, 0)
	return w
}

func (w *world) run(test *models.Test, startup uint) *models.TestRun {
	r := testkit.TestRun(w.db, func(r *models.TestRun) { r.TestID = test.ID })
	w.db.Create(&models.TestRunDeviceStatus{TestRunID: r.ID, DeviceID: w.device.ID, StartupTime: startup})
	return r
}

// protocol stores a protocol of "S/A" on the shared device that measured
// performance.
func (w *world) protocol(r *models.TestRun, fps float64) *models.TestProtocol {
	p := testkit.Protocol(w.db, func(p *models.TestProtocol) {
		p.TestRunID, p.DeviceID, p.TestName, p.AvgFPS = r.ID, &w.device.ID, "S/A", fps
	})
	w.db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "start", Runtime: 1, FPS: fps})
	return p
}

func (w *world) ref(test *models.Test, app *models.App, project *models.Project) testsvc.TestRef {
	return testsvc.TestRef{Scope: testsvc.Scope{ProjectID: project.ID, AppID: app.ID}, TestID: test.ID}
}

func (w *world) key(run *models.TestRun, p *models.TestProtocol) protocol.Key {
	return protocol.Key{TestRef: w.ref(w.mine, w.mineApp, w.mineProject), RunID: run.ID, ProtocolID: p.ID}
}

func assertKind(t *testing.T, err error, want apperr.Kind, message string) {
	t.Helper()
	if got := apperr.KindOf(err); got != want {
		t.Fatalf("error kind %v (%v), want %v", got, err, want)
	}
	if message != "" && err.Error() != message {
		t.Fatalf("message %q, want %q", err, message)
	}
}

// closed is a store whose database is gone.
func closed(t *testing.T) *runstore.Store {
	t.Helper()
	db := testkit.DB(t)
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	return runstore.New(db)
}

func itoa(id uint) string { return strconv.FormatUint(uint64(id), 10) }

// testkitTest stores a test of the own app that has no runs.
func testkitTest(w *world) *models.Test {
	return testkit.Test(w.db, func(x *models.Test) { x.AppID = w.mineApp.ID })
}

func testkitDevice(w *world) *models.Device { return testkit.Device(w.db) }

// protocolWithout stores a protocol of "S/A" on the shared device that
// measured no performance.
func (w *world) protocolWithout(r *models.TestRun) *models.TestProtocol {
	return testkit.Protocol(w.db, func(p *models.TestProtocol) {
		p.TestRunID, p.DeviceID, p.TestName = r.ID, &w.device.ID, "S/A"
	})
}

// protocolOn stores a measuring protocol of "S/A" on the device.
func (w *world) protocolOn(r *models.TestRun, deviceID uint, fps float64) *models.TestProtocol {
	p := testkit.Protocol(w.db, func(p *models.TestProtocol) {
		p.TestRunID, p.DeviceID, p.TestName, p.AvgFPS = r.ID, &deviceID, "S/A", fps
	})
	w.db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "start", FPS: fps})
	return p
}
