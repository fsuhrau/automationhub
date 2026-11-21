package overview_test

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/service/overview"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

func bg() context.Context { return context.Background() }

// fakeRepo answers with its fields and fails the methods in fail.
type fakeRepo struct {
	fail     map[string]error
	totals   overview.Totals
	tests    []models.Test
	last     []models.TestRun
	latest   []models.TestRun
	all      []models.TestProtocol
	failed   []models.TestProtocol
	devices  []models.Device
	testIDs  []uint
	limit    int
	queries  []overview.ProtocolQuery
	totalsOf uint
}

func (f *fakeRepo) Totals(_ context.Context, projectID uint) (overview.Totals, error) {
	f.totalsOf = projectID
	return f.totals, f.fail["Totals"]
}

func (f *fakeRepo) Tests(context.Context, uint) ([]models.Test, error) {
	return f.tests, f.fail["Tests"]
}

func (f *fakeRepo) LastRuns(_ context.Context, ids []uint) ([]models.TestRun, error) {
	f.testIDs = ids
	return f.last, f.fail["LastRuns"]
}

func (f *fakeRepo) LatestRuns(_ context.Context, _ []uint, limit int) ([]models.TestRun, error) {
	f.limit = limit
	return f.latest, f.fail["LatestRuns"]
}

func (f *fakeRepo) Protocols(_ context.Context, q overview.ProtocolQuery) ([]models.TestProtocol, error) {
	f.queries = append(f.queries, q)
	if q.FailedOnly {
		return f.failed, f.fail["FailedProtocols"]
	}
	return f.all, f.fail["Protocols"]
}

func (f *fakeRepo) Devices(context.Context, *models.Project) ([]models.Device, error) {
	return f.devices, f.fail["Devices"]
}

// statusByRun is a summarizer that gives each run the status set for it.
type statusByRun struct {
	status map[uint]string
	calls  int
	failAt int
}

func (s *statusByRun) Summaries(_ context.Context, rs []models.TestRun) ([]testsvc.RunSummary, error) {
	s.calls++
	if s.calls == s.failAt {
		return nil, apperr.Internal("db")
	}
	out := make([]testsvc.RunSummary, len(rs))
	for i := range rs {
		out[i] = testsvc.RunSummary{ID: rs[i].ID, Status: s.status[rs[i].ID]}
	}
	return out, nil
}

type env struct {
	svc     *overview.Service
	repo    *fakeRepo
	summary *statusByRun
	pool    *testkit.FakeDeviceManager
}

func newEnv() *env {
	e := &env{
		repo:    &fakeRepo{fail: map[string]error{}},
		summary: &statusByRun{status: map[uint]string{}},
		pool:    &testkit.FakeDeviceManager{},
	}
	e.svc = overview.New(overview.Deps{Repo: e.repo, Summarizer: e.summary, Pool: e.pool})
	return e
}

func assertKind(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	if got := apperr.KindOf(err); got != want {
		t.Fatalf("error kind %v (%v), want %v", got, err, want)
	}
}

func test(id, appID uint, name string) models.Test {
	return models.Test{Model: models.Model{ID: id}, AppID: appID, Name: name}
}

func runOf(id, testID uint) models.TestRun {
	return models.TestRun{Model: models.Model{ID: id}, TestID: testID}
}

func TestGetSortsTheLatestRunOfEachTestIntoFailingAndRunning(t *testing.T) {
	e := newEnv()
	e.repo.totals = overview.Totals{Apps: 2, Bundles: 3, BundleStorage: 150}
	e.repo.tests = []models.Test{test(1, 10, "smoke"), test(2, 10, "payment"), test(3, 11, "nightly"), test(4, 11, "fresh")}
	e.repo.last = []models.TestRun{runOf(40, 1), runOf(30, 2), runOf(20, 3), runOf(15, 4), runOf(10, 5)}
	e.repo.latest = []models.TestRun{runOf(40, 1), runOf(30, 2)}
	e.summary.status = map[uint]string{40: "failed", 30: "passed", 20: "running", 15: "not_run", 10: "unstable"}

	o, err := e.svc.Get(bg(), 7)

	if err != nil {
		t.Fatal(err)
	}
	if o.Apps != 2 || o.Tests != 4 || o.Bundles != 3 || o.BundleStorage != 150 || e.repo.totalsOf != 7 {
		t.Errorf("counts %+v", o)
	}
	if len(o.Failing) != 3 || o.Failing[0].ID != 40 || o.Failing[0].TestName != "smoke" || o.Failing[0].AppID != 10 {
		t.Errorf("failing %+v", o.Failing)
	}
	if len(o.Running) != 1 || o.Running[0].ID != 20 || o.Running[0].TestName != "nightly" {
		t.Errorf("running %+v", o.Running)
	}
	if len(o.Latest) != 2 || o.Latest[1].TestName != "payment" || e.repo.limit != overview.LatestRuns {
		t.Errorf("latest %+v, limit %d", o.Latest, e.repo.limit)
	}
	if len(e.repo.testIDs) != 4 {
		t.Errorf("last runs asked for tests %v", e.repo.testIDs)
	}
}

func TestGetOfAProjectWithoutTestsHasEmptyListsAndAsksForNoRuns(t *testing.T) {
	e := newEnv()
	e.repo.totals = overview.Totals{Apps: 1}
	e.repo.fail["LastRuns"] = apperr.Internal("must not be asked")

	o, err := e.svc.Get(bg(), 7)

	if err != nil || o.Apps != 1 || o.Failing == nil || o.Running == nil || o.Latest == nil || len(o.Latest) != 0 {
		t.Fatalf("overview %+v, err %v", o, err)
	}
}

func TestGetErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*env)
	}{
		{"totals", func(e *env) { e.repo.fail["Totals"] = apperr.Internal("db") }},
		{"tests", func(e *env) { e.repo.fail["Tests"] = apperr.Internal("db") }},
		{"last runs", func(e *env) { e.repo.fail["LastRuns"] = apperr.Internal("db") }},
		{"last summaries", func(e *env) { e.summary.failAt = 1 }},
		{"latest runs", func(e *env) { e.repo.fail["LatestRuns"] = apperr.Internal("db") }},
		{"latest summaries", func(e *env) { e.summary.failAt = 2 }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv()
			e.repo.tests = []models.Test{test(1, 10, "smoke")}
			e.repo.last = []models.TestRun{runOf(40, 1)}
			e.repo.latest = []models.TestRun{runOf(40, 1)}
			tc.setup(e)

			o, err := e.svc.Get(bg(), 7)

			assertKind(t, err, apperr.KindInternal)
			if o != nil {
				t.Fatalf("overview %+v with an error", o)
			}
		})
	}
}

func TestStatsCountsBundlesAndBootedDevicesAndListsProtocols(t *testing.T) {
	e := newEnv()
	e.repo.totals = overview.Totals{Bundles: 3, BundleStorage: 150}
	e.repo.all = []models.TestProtocol{{TestName: "a"}, {TestName: "b"}}
	e.repo.failed = []models.TestProtocol{{TestName: "b"}}
	e.repo.devices = []models.Device{{DeviceIdentifier: "booted"}, {DeviceIdentifier: "off"}, {DeviceIdentifier: "gone"}}
	booted, off := testkit.NewFakeDevice("booted"), testkit.NewFakeDevice("off")
	booted.State, off.State = device.StateBooted, device.StateShutdown
	e.pool.Add("android_device", booted)
	e.pool.Add("android_device", off)

	stats, err := e.svc.Stats(bg(), &models.Project{Model: models.Model{ID: 7}})

	if err != nil {
		t.Fatal(err)
	}
	if stats.AppsCount != 3 || stats.AppsStorageSize != 150 || stats.DeviceCount != 3 || stats.DeviceBooted != 1 {
		t.Errorf("stats %+v", stats)
	}
	if len(stats.TestsLastProtocols) != 2 || len(stats.TestsLastFailed) != 1 {
		t.Errorf("protocols %d, failed %d", len(stats.TestsLastProtocols), len(stats.TestsLastFailed))
	}
	if len(e.repo.queries) != 2 || e.repo.queries[0].ProjectID != 7 || e.repo.queries[0].FailedOnly ||
		!e.repo.queries[1].FailedOnly || e.repo.queries[1].Limit != 10 {
		t.Errorf("protocol queries %+v", e.repo.queries)
	}
}

func TestStatsErrors(t *testing.T) {
	for _, fail := range []string{"Totals", "Protocols", "FailedProtocols", "Devices"} {
		t.Run(fail, func(t *testing.T) {
			e := newEnv()
			e.repo.fail[fail] = apperr.Internal("db")

			stats, err := e.svc.Stats(bg(), &models.Project{Model: models.Model{ID: 7}})

			assertKind(t, err, apperr.KindInternal)
			if stats != nil {
				t.Fatalf("stats %+v with an error", stats)
			}
		})
	}
}
