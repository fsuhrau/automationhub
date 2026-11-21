// Package runfixture is the test setup of the run, protocol, overview and
// stats routes: the api with two projects that have one app, test and run
// each. The foreign project is created first, so its ids are the lower ones
// and an unscoped "earlier runs" query would pick them up.
package runfixture

import (
	"fmt"
	"testing"

	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

// World is the api with the projects "foreign" and "mine".
type World struct {
	*apitest.Env
	Device                    *models.Device
	MineApp, ForeignApp       *models.App
	Mine, Foreign             *models.Test
	Run1, Run2, ForeignRun    *models.TestRun
	Proto1, Proto2, ForeProto *models.TestProtocol
}

// New builds the world: the foreign run has a protocol of 100 fps and a
// startup time of 1000, mine has two runs with 10 and 0 fps and startup
// times of 20 and 30. All protocols measured performance.
func New(t testing.TB, opts ...apitest.Option) *World {
	t.Helper()
	w := &World{Env: apitest.New(t, opts...)}
	db := w.DB
	w.Device = testkit.Device(db)

	foreign := testkit.Project(db, func(p *models.Project) { p.Identifier = "foreign" })
	w.ForeignApp = testkit.App(db, func(a *models.App) { a.ProjectID, a.Name = foreign.ID, "foreign" })
	w.Foreign = testkit.Test(db, func(x *models.Test) { x.AppID, x.Name = w.ForeignApp.ID, "foreign" })
	w.ForeignRun = w.run(w.Foreign, 1000)
	w.ForeProto = w.Protocol(w.ForeignRun, 100)

	mine := testkit.Project(db, func(p *models.Project) { p.Identifier = "mine" })
	w.MineApp = testkit.App(db, func(a *models.App) { a.ProjectID, a.Name = mine.ID, "mine" })
	w.Mine = testkit.Test(db, func(x *models.Test) { x.AppID, x.Name = w.MineApp.ID, "smoke" })
	w.Run1 = w.run(w.Mine, 20)
	w.Proto1 = w.Protocol(w.Run1, 10)
	w.Run2 = w.run(w.Mine, 30)
	w.Proto2 = w.Protocol(w.Run2, 0)
	return w
}

// run stores a run of the test whose device started in startup.
func (w *World) run(test *models.Test, startup uint) *models.TestRun {
	r := testkit.TestRun(w.DB, func(r *models.TestRun) { r.TestID = test.ID })
	w.DB.Create(&models.TestRunDeviceStatus{TestRunID: r.ID, DeviceID: w.Device.ID, StartupTime: startup})
	return r
}

// Protocol stores a protocol of "S/A" on the shared device that measured
// performance, so it counts for histories.
func (w *World) Protocol(r *models.TestRun, fps float64) *models.TestProtocol {
	p := testkit.Protocol(w.DB, func(p *models.TestProtocol) {
		p.TestRunID, p.DeviceID, p.TestName, p.AvgFPS = r.ID, &w.Device.ID, "S/A", fps
	})
	w.DB.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "start", Runtime: 1, FPS: fps})
	return p
}

// Path is a path below the app of the project "mine".
func (w *World) Path(format string, args ...any) string {
	return fmt.Sprintf("/api/mine/app/%d", w.MineApp.ID) + fmt.Sprintf(format, args...)
}
