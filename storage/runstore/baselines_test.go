package runstore_test

import (
	"fmt"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/protocol"
	"github.com/fsuhrau/automationhub/storage/models"
)

// baselineWorld is the test "perf" with a current protocol and candidates
// on bundles with tags, on two devices and in another app.
type baselineWorld struct {
	*world
	phone, tablet                          *models.Device
	onRelease, onCandidate, onTablet       *models.TestProtocol
	unmeasured, inOtherApp, current, later *models.TestProtocol
}

func newBaselineWorld(t *testing.T) *baselineWorld {
	t.Helper()
	w := &baselineWorld{world: newWorld(t)}
	w.phone, w.tablet = w.device, testkitDevice(w.world)
	release := w.bundle("1.0", "release,store")
	candidate := w.bundle("1.1", "release_candidate")
	nightly := w.bundle("1.2", "nightly")
	w.onRelease = w.runOf(w.mine, release, w.phone, 60, true)
	w.onCandidate = w.runOf(w.mine, candidate, w.phone, 55, true)
	w.onTablet = w.runOf(w.mine, release, w.tablet, 30, true)
	w.unmeasured = w.runOf(w.mine, release, w.phone, 0, false)
	w.inOtherApp = w.runOf(w.foreign, release, w.phone, 60, true)
	w.current = w.runOf(w.mine, nightly, w.phone, 40, true)
	w.later = w.runOf(w.mine, nightly, w.phone, 41, true)
	return w
}

func (w *baselineWorld) bundle(version, tags string) *models.AppBinary {
	b := &models.AppBinary{AppID: w.mineApp.ID, Name: "Game", Version: version, Tags: tags}
	w.db.Create(b)
	return b
}

// runOf adds a run of the test on the bundle with one protocol of "Suite/Perf".
func (w *baselineWorld) runOf(test *models.Test, b *models.AppBinary, d *models.Device, fps float64, measured bool) *models.TestProtocol {
	r := &models.TestRun{TestID: test.ID, AppBinaryID: b.ID, Parameter: "Level=" + b.Version}
	w.db.Create(r)
	p := &models.TestProtocol{TestRunID: r.ID, DeviceID: &d.ID, TestName: "Suite/Perf", AvgFPS: fps}
	w.db.Create(p)
	if measured {
		w.db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "start", Runtime: 1, FPS: fps})
		w.db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "end", Runtime: 3, FPS: fps})
	}
	return p
}

func ids(found []models.TestProtocol) string {
	out := []uint{}
	for _, p := range found {
		out = append(out, p.ID)
	}
	return fmt.Sprint(out)
}

func (w *baselineWorld) candidates(t *testing.T, f protocol.Filter) []models.TestProtocol {
	t.Helper()
	if f.Limit == 0 {
		f.Limit = 50
	}
	got, err := w.store.Candidates(bg(), protocol.CandidateQuery{AppID: w.mineApp.ID, Current: w.current, Filter: f})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestCandidatesAreMeasuredProtocolsOfTheSameTestNameInTheApp(t *testing.T) {
	w := newBaselineWorld(t)
	tests := []struct {
		name   string
		filter protocol.Filter
		want   []*models.TestProtocol
	}{
		{"same device", protocol.Filter{}, []*models.TestProtocol{w.later, w.onCandidate, w.onRelease}},
		{"any device", protocol.Filter{AnyDevice: true}, []*models.TestProtocol{w.later, w.onTablet, w.onCandidate, w.onRelease}},
		{"tag", protocol.Filter{Tag: "release"}, []*models.TestProtocol{w.onRelease}},
		{"tag matches exactly", protocol.Filter{Tag: "release_candidate"}, []*models.TestProtocol{w.onCandidate}},
		{"tag on any device", protocol.Filter{Tag: "release", AnyDevice: true}, []*models.TestProtocol{w.onTablet, w.onRelease}},
		{"text in the version", protocol.Filter{Query: "1.1"}, []*models.TestProtocol{w.onCandidate}},
		{"text in the tags", protocol.Filter{Query: "candidate"}, []*models.TestProtocol{w.onCandidate}},
		{"text in the run parameters", protocol.Filter{Query: "level=1.0"}, []*models.TestProtocol{w.onRelease}},
		{"run id", protocol.Filter{Query: fmt.Sprintf("#%d", w.onRelease.TestRunID)}, []*models.TestProtocol{w.onRelease}},
		{"underscore matches literally", protocol.Filter{Query: "1_1"}, nil},
		{"limit", protocol.Filter{Limit: 1}, []*models.TestProtocol{w.later}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := w.candidates(t, tc.filter)

			want := []models.TestProtocol{}
			for _, p := range tc.want {
				want = append(want, *p)
			}
			if ids(got) != ids(want) {
				t.Errorf("got %s, want %s", ids(got), ids(want))
			}
		})
	}
}

func TestCandidatesCarryDeviceBundleAndRun(t *testing.T) {
	w := newBaselineWorld(t)

	got := w.candidates(t, protocol.Filter{})

	last := got[len(got)-1]
	if last.Device == nil || last.TestRun == nil || last.TestRun.AppBinary == nil || last.TestRun.AppBinary.Version != "1.0" {
		t.Fatalf("candidate %+v", last)
	}
}

func TestCandidatesOfAProtocolWithoutADeviceAreThoseWithoutOne(t *testing.T) {
	w := newBaselineWorld(t)
	noDevice := &models.TestProtocol{TestRunID: w.onRelease.TestRunID, TestName: "Suite/Perf"}
	w.db.Create(noDevice)
	w.db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: noDevice.ID, Checkpoint: "start"})
	w.current = &models.TestProtocol{TestRunID: w.later.TestRunID, TestName: "Suite/Perf"}

	got := w.candidates(t, protocol.Filter{})

	if ids(got) != ids([]models.TestProtocol{*noDevice}) {
		t.Fatalf("got %s, want only the protocol without a device %d", ids(got), noDevice.ID)
	}
}

func TestBaselineLoadsTheCandidateOfARunWithItsCheckpointsInOrder(t *testing.T) {
	w := newBaselineWorld(t)
	q := protocol.BaselineQuery{AppID: w.mineApp.ID, Current: w.current, RunID: w.onTablet.TestRunID}

	got, err := w.store.Baseline(bg(), q)

	if err != nil || got.ID != w.onTablet.ID || len(got.Performance) != 2 || got.Performance[0].Checkpoint != "start" ||
		got.Device == nil || got.TestRun.AppBinary == nil {
		t.Fatalf("baseline %+v, err %v", got, err)
	}
}

func TestBaselinePrefersTheSameDeviceUnlessAnotherIsAskedFor(t *testing.T) {
	w := newBaselineWorld(t)
	both := &models.TestRun{TestID: w.mine.ID}
	w.db.Create(both)
	for _, d := range []*models.Device{w.tablet, w.phone} {
		p := &models.TestProtocol{TestRunID: both.ID, DeviceID: &d.ID, TestName: "Suite/Perf"}
		w.db.Create(p)
		w.db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "start"})
	}
	q := protocol.BaselineQuery{AppID: w.mineApp.ID, Current: w.current, RunID: both.ID}

	same, err := w.store.Baseline(bg(), q)
	if err != nil || *same.DeviceID != w.phone.ID {
		t.Fatalf("baseline %+v, err %v; want the current protocol's device", same, err)
	}
	q.DeviceID = &w.tablet.ID
	asked, err := w.store.Baseline(bg(), q)
	if err != nil || *asked.DeviceID != w.tablet.ID {
		t.Fatalf("baseline %+v, err %v; want the asked for device", asked, err)
	}
}

func TestBaselineOfARunThatCannotBeComparedWithIsNotFound(t *testing.T) {
	w := newBaselineWorld(t)
	for name, p := range map[string]*models.TestProtocol{"unmeasured": w.unmeasured, "other app": w.inOtherApp, "the run itself": w.current} {
		t.Run(name, func(t *testing.T) {
			q := protocol.BaselineQuery{AppID: w.mineApp.ID, Current: w.current, RunID: p.TestRunID}

			_, err := w.store.Baseline(bg(), q)

			assertKind(t, err, apperr.KindNotFound, fmt.Sprintf("run %d has no performance of Suite/Perf to compare with", p.TestRunID))
		})
	}
}

func TestBundleTagsListsTheDistinctTagListsOfTheApp(t *testing.T) {
	w := newBaselineWorld(t)

	got, err := w.store.BundleTags(bg(), w.mineApp.ID)

	if err != nil || len(got) != 3 {
		t.Fatalf("tags %v, err %v; want the 3 distinct lists", got, err)
	}
}

func TestBaselineReadsReportADatabaseError(t *testing.T) {
	s := closed(t)
	current := &models.TestProtocol{}
	_, e1 := s.Candidates(bg(), protocol.CandidateQuery{Current: current})
	_, e2 := s.Baseline(bg(), protocol.BaselineQuery{Current: current})
	_, e3 := s.BundleTags(bg(), 1)
	for i, err := range []error{e1, e2, e3} {
		if apperr.KindOf(err) != apperr.KindInternal {
			t.Errorf("read %d: %v", i, err)
		}
	}
}
