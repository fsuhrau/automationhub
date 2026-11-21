package protocols_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/api/internal/runfixture"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/service/protocol"
	"github.com/fsuhrau/automationhub/storage/models"
)

// The protocol page compares a test with its earlier runs. Only earlier runs
// on the same device that measured performance count: not the protocol
// itself, not its runner (same test name, no measurements), not later runs.
func TestProtocolHistoryAverages(t *testing.T) {
	w := runfixture.New(t)
	other := testkit.Device(w.DB)
	test := w.Mine.ID
	measuring := func(run uint, parent *uint, deviceID uint, fps float64, measured bool) uint {
		p := testkit.Protocol(w.DB, func(p *models.TestProtocol) {
			p.TestRunID, p.ParentTestProtocolID, p.DeviceID, p.TestName, p.AvgFPS = run, parent, &deviceID, "Smoke/Main", fps
		})
		if measured {
			w.DB.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "Load", FPS: fps})
		}
		return p.ID
	}
	var runs []uint
	for _, fps := range []float64{10, 30, 50, 100} {
		r := testkit.TestRun(w.DB, func(r *models.TestRun) { r.TestID = test })
		runs = append(runs, r.ID)
		runner := measuring(r.ID, nil, w.Device.ID, 0, false)
		measuring(r.ID, &runner, w.Device.ID, fps, true)
	}
	measuring(runs[1], nil, other.ID, 999, true) // another device
	var current models.TestProtocol
	w.DB.Where("test_run_id = ? and parent_test_protocol_id is not null", runs[2]).First(&current)

	got := testkit.Decode[models.TestRun](t, w.Do("GET", w.Path("/test/%d/run/%d/%d", test, runs[2], current.ID), nil))

	if len(got.Protocols) != 1 || got.Protocols[0].ID != current.ID {
		t.Fatalf("protocols %+v", got.Protocols)
	}
	p := got.Protocols[0]
	if p.HistAvgFPS != 20 {
		t.Errorf("hist avg fps %v, want 20 (runs 1 and 2)", p.HistAvgFPS)
	}
	var history []uint
	for _, h := range p.TestProtocolHistory {
		history = append(history, h.TestRunID)
	}
	if fmt.Sprint(history) != fmt.Sprint([]uint{runs[1], runs[0]}) {
		t.Errorf("history runs %v, want [%d %d]", history, runs[1], runs[0])
	}
	var first models.TestProtocol
	w.DB.Where("test_run_id = ? and parent_test_protocol_id is not null", runs[0]).First(&first)
	earliest := testkit.Decode[models.TestRun](t, w.Do("GET", w.Path("/test/%d/run/%d/%d", test, runs[0], first.ID), nil))
	if hist := earliest.Protocols[0]; hist.HistAvgFPS != 0 || len(hist.TestProtocolHistory) != 0 {
		t.Errorf("first run: hist %v, history %d", hist.HistAvgFPS, len(hist.TestProtocolHistory))
	}
}

// baselines is the test "Suite/Perf" with a current protocol and runs on
// bundles with tags, on two devices and in another app.
type baselines struct {
	*runfixture.World
	phone, tablet                          *models.Device
	onRelease, onCandidate, onTablet       *models.TestProtocol
	unmeasured, inOtherApp, current, later *models.TestProtocol
	release                                *models.AppBinary
}

func newBaselines(t *testing.T) *baselines {
	t.Helper()
	b := &baselines{World: runfixture.New(t)}
	b.phone = testkit.Device(b.DB, func(d *models.Device) { d.Name = "phone" })
	b.tablet = testkit.Device(b.DB, func(d *models.Device) { d.Name, d.Alias = "tablet", "Tab" })
	b.release = b.bundle("1.0", "release,store")
	candidate := b.bundle("1.1", "release_candidate")
	nightly := b.bundle("1.2", "nightly")
	b.onRelease = b.perf(b.Mine, b.release, b.phone, 60, true)
	b.onCandidate = b.perf(b.Mine, candidate, b.phone, 55, true)
	b.onTablet = b.perf(b.Mine, b.release, b.tablet, 30, true)
	b.unmeasured = b.perf(b.Mine, b.release, b.phone, 0, false)
	b.inOtherApp = b.perf(b.Foreign, b.release, b.phone, 60, true)
	b.current = b.perf(b.Mine, nightly, b.phone, 40, true)
	b.later = b.perf(b.Mine, nightly, b.phone, 41, true)
	return b
}

func (b *baselines) bundle(version, tags string) *models.AppBinary {
	bundle := &models.AppBinary{AppID: b.MineApp.ID, Name: "Game", Version: version, Tags: tags}
	b.DB.Create(bundle)
	return bundle
}

func (b *baselines) perf(test *models.Test, bundle *models.AppBinary, d *models.Device, fps float64, measured bool) *models.TestProtocol {
	r := testkit.TestRun(b.DB, func(r *models.TestRun) { r.TestID, r.AppBinaryID = test.ID, bundle.ID })
	p := testkit.Protocol(b.DB, func(p *models.TestProtocol) {
		p.TestRunID, p.DeviceID, p.TestName, p.AvgFPS = r.ID, &d.ID, "Suite/Perf", fps
	})
	if measured {
		b.DB.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "start", Runtime: 1, FPS: fps})
		b.DB.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "end", Runtime: 3, FPS: fps})
	}
	return p
}

func (b *baselines) base() string {
	return b.Path("/test/%d/run/%d/%d", b.Mine.ID, b.current.TestRunID, b.current.ID)
}

func (b *baselines) list(t *testing.T, query string) protocol.Baselines {
	t.Helper()
	rec := b.Do("GET", b.base()+"/baselines"+query, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s: %d %s", query, rec.Code, rec.Body)
	}
	return testkit.Decode[protocol.Baselines](t, rec)
}

// A protocol's performance can be compared with another run of the same test
// picked by bundle name, version, tag or run id, on the same device or on
// any; runs without performance, other tests and other apps aren't offered.
func TestBaselinesOfAProtocol(t *testing.T) {
	b := newBaselines(t)
	expect := func(query string, want ...*models.TestProtocol) {
		t.Helper()
		got := []uint{}
		for _, c := range b.list(t, query).Candidates {
			got = append(got, c.ProtocolID)
		}
		ids := []uint{}
		for _, p := range want {
			ids = append(ids, p.ID)
		}
		if fmt.Sprint(got) != fmt.Sprint(ids) {
			t.Errorf("%q: got %v, want %v", query, got, ids)
		}
	}

	expect("", b.later, b.onCandidate, b.onRelease)
	expect("?anyDevice=true", b.later, b.onTablet, b.onCandidate, b.onRelease)
	expect("?tag=release", b.onRelease)
	expect("?tag=release_candidate", b.onCandidate)
	expect("?tag=release&anyDevice=true", b.onTablet, b.onRelease)
	expect("?q=1.1", b.onCandidate)
	expect("?q=CANDIDATE", b.onCandidate)
	expect(fmt.Sprintf("?q=%%23%d", b.onRelease.TestRunID), b.onRelease)
	expect("?q=1_1")
	expect("?limit=1", b.later)

	all := b.list(t, "")
	if fmt.Sprint(all.Tags) != "[nightly release release_candidate store]" {
		t.Errorf("tags %v", all.Tags)
	}
	if c := all.Candidates[len(all.Candidates)-1]; c.Bundle == nil || c.Bundle.Version != "1.0" || c.FPS != 60 || c.Device != "phone" {
		t.Errorf("oldest candidate %+v", c)
	}
}

func TestBaselineLoadsOneRunWithItsCheckpoints(t *testing.T) {
	b := newBaselines(t)

	rec := b.Do("GET", fmt.Sprintf("%s/baseline/%d", b.base(), b.onTablet.TestRunID), nil)

	got := testkit.Decode[protocol.Baseline](t, rec)
	if rec.Code != http.StatusOK || got.ProtocolID != b.onTablet.ID || got.Device != "Tab" || len(got.Performance) != 2 || got.Bundle.Tags != "release,store" {
		t.Fatalf("baseline: %d %+v", rec.Code, got)
	}
	if got.Performance[0].Checkpoint != "start" {
		t.Errorf("checkpoints out of order: %+v", got.Performance)
	}
}

// A run with the test on several devices: the same device unless another is
// asked for.
func TestBaselinePrefersTheSameDevice(t *testing.T) {
	b := newBaselines(t)
	both := testkit.TestRun(b.DB, func(r *models.TestRun) { r.TestID, r.AppBinaryID = b.Mine.ID, b.release.ID })
	for _, d := range []*models.Device{b.tablet, b.phone} {
		p := testkit.Protocol(b.DB, func(p *models.TestProtocol) { p.TestRunID, p.DeviceID, p.TestName = both.ID, &d.ID, "Suite/Perf" })
		b.DB.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "start", Runtime: 1})
	}
	deviceOf := func(query string) uint {
		rec := b.Do("GET", fmt.Sprintf("%s/baseline/%d%s", b.base(), both.ID, query), nil)
		return *testkit.Decode[protocol.Baseline](t, rec).DeviceID
	}

	if got := deviceOf(""); got != b.phone.ID {
		t.Errorf("baseline on device %d, want the same device %d", got, b.phone.ID)
	}
	if got := deviceOf(fmt.Sprintf("?deviceId=%d", b.tablet.ID)); got != b.tablet.ID {
		t.Errorf("baseline on device %d, want the asked for %d", got, b.tablet.ID)
	}
}

func TestBaselineOfARunThatCannotBeComparedWithIsNotFound(t *testing.T) {
	b := newBaselines(t)
	for name, p := range map[string]*models.TestProtocol{"unmeasured": b.unmeasured, "other app": b.inOtherApp, "the run itself": b.current} {
		t.Run(name, func(t *testing.T) {
			rec := b.Do("GET", fmt.Sprintf("%s/baseline/%d", b.base(), p.TestRunID), nil)

			if rec.Code != http.StatusNotFound {
				t.Errorf("run %d offered as baseline: %d", p.TestRunID, rec.Code)
			}
		})
	}
}

// The protocol has to be in the run and the run in the test of the app.
func TestBaselinesRefuseAProtocolOutsideTheRunAndTest(t *testing.T) {
	b := newBaselines(t)
	for _, path := range []string{
		b.Path("/test/%d/run/%d/%d/baselines", b.Mine.ID, b.onRelease.TestRunID, b.current.ID),
		b.Path("/test/%d/run/%d/%d/baselines", b.Foreign.ID, b.inOtherApp.TestRunID, b.inOtherApp.ID),
	} {
		if rec := b.Do("GET", path, nil); rec.Code != http.StatusNotFound {
			t.Errorf("%s: %d", path, rec.Code)
		}
	}
}
