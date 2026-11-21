package run_test

import (
	"context"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/service/run"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

func bg() context.Context { return context.Background() }

// fakeRepo is an in-memory run.Repo: it answers with its fields, fails the
// methods in fail, and records what it was asked.
type fakeRepo struct {
	fail map[string]error

	test       *models.Test
	bundle     *models.AppBinary
	devices    []models.Device
	testRuns   []models.TestRun
	ends       []run.RunEnd
	lastRun    *models.TestRun
	runDetail  *models.TestRun
	around     run.Neighbours
	startups   []models.TestRunDeviceStatus
	protocol   *models.TestProtocol
	logs       []models.ProtocolEntry
	logsTotal  int64
	logQuery   run.LogQuery
	deviceQ    run.DeviceQuery
	before     uint
	limit      int
	runIDsSeen []uint

	reportRun *models.TestRun
	logTail   []models.TestRunLogEntry
	tailLimit int
	progress  run.Progress
	digest    run.Digest
	dataList  []models.ProtocolEntry
	digestFor uint
	digestMax int
}

func (f *fakeRepo) err(method string) error { return f.fail[method] }

func (f *fakeRepo) StartTest(context.Context, uint, uint) (*models.Test, error) {
	return f.test, f.err("StartTest")
}

func (f *fakeRepo) Bundle(context.Context, uint, uint) (*models.AppBinary, error) {
	return f.bundle, f.err("Bundle")
}

func (f *fakeRepo) Devices(_ context.Context, q run.DeviceQuery) ([]models.Device, error) {
	f.deviceQ = q
	return f.devices, f.err("Devices")
}

func (f *fakeRepo) RunInProject(context.Context, uint, uint) error { return f.err("RunInProject") }

func (f *fakeRepo) TestInScope(context.Context, testsvc.TestRef) error { return f.err("TestInScope") }

func (f *fakeRepo) RunInTest(context.Context, uint, uint) error { return f.err("RunInTest") }

func (f *fakeRepo) TestRuns(_ context.Context, q run.PageQuery) ([]models.TestRun, error) {
	f.before, f.limit = q.Before, q.Limit
	return f.testRuns, f.err("TestRuns")
}

func (f *fakeRepo) RunEnds(_ context.Context, ids []uint) ([]run.RunEnd, error) {
	f.runIDsSeen = ids
	return f.ends, f.err("RunEnds")
}

func (f *fakeRepo) LastRun(context.Context, uint) (*models.TestRun, error) {
	return f.lastRun, f.err("LastRun")
}

func (f *fakeRepo) Run(context.Context, uint) (*models.TestRun, error) {
	return f.runDetail, f.err("Run")
}

func (f *fakeRepo) Neighbours(context.Context, uint, uint) (run.Neighbours, error) {
	return f.around, f.err("Neighbours")
}

func (f *fakeRepo) StartupTimes(context.Context, run.StartupQuery) ([]models.TestRunDeviceStatus, error) {
	return f.startups, f.err("StartupTimes")
}

func (f *fakeRepo) Device(context.Context, *models.Project, uint) (*models.Device, error) {
	return &f.devices[0], f.err("Device")
}

func (f *fakeRepo) ProjectProtocol(context.Context, uint, uint) (*models.TestProtocol, error) {
	return f.protocol, f.err("ProjectProtocol")
}

func (f *fakeRepo) ProtocolLogs(_ context.Context, q run.LogQuery) ([]models.ProtocolEntry, int64, error) {
	f.logQuery = q
	return f.logs, f.logsTotal, f.err("ProtocolLogs")
}

func (f *fakeRepo) ReportRun(context.Context, uint) (*models.TestRun, error) {
	return f.reportRun, f.err("ReportRun")
}

func (f *fakeRepo) LogTail(_ context.Context, _ uint, limit int) ([]models.TestRunLogEntry, error) {
	f.tailLimit = limit
	return f.logTail, f.err("LogTail")
}

func (f *fakeRepo) ProtocolProgress(context.Context, uint) (run.Progress, error) {
	return f.progress, f.err("ProtocolProgress")
}

func (f *fakeRepo) ProtocolDigest(_ context.Context, protocolID uint, limit int) (run.Digest, error) {
	f.digestFor, f.digestMax = protocolID, limit
	return f.digest, f.err("ProtocolDigest")
}

func (f *fakeRepo) DataEntries(context.Context, uint) ([]models.ProtocolEntry, error) {
	return f.dataList, f.err("DataEntries")
}

// fakeRunner is a run.Runner that records its calls.
type fakeRunner struct {
	failOn   string
	err      error
	env      map[string]string
	test     models.Test
	devices  []models.Device
	binary   *models.AppBinary
	startURL string
	cancelID string
	held     []uint
	started  *models.TestRun
}

func (r *fakeRunner) fails(method string) error {
	if r.failOn == method {
		return r.err
	}
	return nil
}

func (r *fakeRunner) Initialize(test models.Test, env map[string]string) error {
	r.test, r.env = test, env
	return r.fails("Initialize")
}

func (r *fakeRunner) Run(devs []models.Device, binary *models.AppBinary, startURL string) (*models.TestRun, error) {
	r.devices, r.binary, r.startURL = devs, binary, startURL
	return r.started, r.fails("Run")
}

func (r *fakeRunner) Cancel(runID string) error {
	r.cancelID = runID
	return r.fails("Cancel")
}

func (r *fakeRunner) LockedDevices() []uint { return r.held }

// registry is an in-memory run.Runners.
type registry map[uint]run.Runner

func (r registry) Add(id uint, runner run.Runner) { r[id] = runner }
func (r registry) Get(id uint) (run.Runner, bool) { runner, ok := r[id]; return runner, ok }
func (r registry) Remove(id uint)                 { delete(r, id) }

// summarizer gives every run the status "passed".
type summarizer struct{ err error }

func (s summarizer) Summaries(_ context.Context, rs []models.TestRun) ([]testsvc.RunSummary, error) {
	out := make([]testsvc.RunSummary, len(rs))
	for i := range rs {
		out[i] = testsvc.RunSummary{ID: rs[i].ID, Status: "passed"}
	}
	return out, s.err
}

type env struct {
	svc      *run.Service
	repo     *fakeRepo
	pool     *testkit.FakeDeviceManager
	runners  registry
	runner   *fakeRunner
	summary  *summarizer
	tracked  []uint
	kinds    []models.TestType
	runState map[uint]runs.RunState
	project  *models.Project
	newErr   error
}

func newEnv() *env {
	e := &env{
		repo:     &fakeRepo{fail: map[string]error{}},
		pool:     &testkit.FakeDeviceManager{},
		runners:  registry{},
		runner:   &fakeRunner{started: &models.TestRun{Model: models.Model{ID: 5}}},
		summary:  &summarizer{},
		runState: map[uint]runs.RunState{},
		project:  &models.Project{Model: models.Model{ID: 1}, Identifier: "proj"},
	}
	e.svc = run.New(run.Deps{
		Repo: e.repo, Pool: e.pool, Runners: e.runners, Summarizer: e.summary,
		NewRunner: func(kind models.TestType, _ *models.Project, _ uint) (run.Runner, error) {
			e.kinds = append(e.kinds, kind)
			return e.runner, e.newErr
		},
		RunState: func(id uint) runs.RunState { return e.runState[id] },
		Started:  func(id uint) { e.tracked = append(e.tracked, id) },
	})
	return e
}

func ref() testsvc.TestRef {
	return testsvc.TestRef{Scope: testsvc.Scope{ProjectID: 1, AppID: 2}, TestID: 3}
}

func assertKind(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	if got := apperr.KindOf(err); got != want {
		t.Fatalf("error kind %v (%v), want %v", got, err, want)
	}
}

func at(sec int) time.Time { return time.Unix(int64(sec), 0).UTC() }
