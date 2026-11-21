package control

import (
	"errors"
	"net/http"
	"testing"

	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestLeaseFlow(t *testing.T) {
	e := setup(t)
	dev := e.full
	tap := map[string]int{"x": 1, "y": 2}

	if rec := e.call(t, "", http.MethodPost, "1/tap", tap); rec.Code != http.StatusPreconditionRequired {
		t.Fatalf("tap without lease: expected 428, got %d %s", rec.Code, rec.Body)
	}
	rec := testkit.DoJSON(t, as(e.router, "alice", ""), http.MethodPost, "/api/p/device/1/lock", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("lock: %d %s", rec.Code, rec.Body)
	}
	lease := testkit.Decode[hubcontrol.Lease](t, rec)
	if lease.Token == "" || lease.Owner != "alice" {
		t.Fatalf("unexpected lease %+v", lease)
	}

	if rec := testkit.DoJSON(t, as(e.router, "bob", ""), http.MethodPost, "/api/p/device/1/lock", nil); rec.Code != http.StatusConflict {
		t.Fatalf("second owner: expected 409, got %d", rec.Code)
	}
	if rec := e.call(t, "bogus", http.MethodPost, "1/tap", tap); rec.Code != http.StatusConflict {
		t.Fatalf("wrong token: expected 409, got %d", rec.Code)
	}
	if rec := e.call(t, lease.Token, http.MethodPost, "1/tap", map[string]int{}); rec.Code != http.StatusBadRequest {
		t.Fatalf("tap without coordinates: expected 400, got %d", rec.Code)
	}
	if rec := e.call(t, lease.Token, http.MethodPost, "1/tap", map[string]int{"x": 10, "y": 20}); rec.Code != http.StatusNoContent {
		t.Fatalf("tap: expected 204, got %d %s", rec.Code, rec.Body)
	}
	if rec := e.call(t, lease.Token, http.MethodPost, "1/key", map[string]string{"key": "back"}); rec.Code != http.StatusNoContent {
		t.Fatalf("key: expected 204, got %d", rec.Code)
	}
	if taps := dev.CallsTo("Tap"); len(taps) != 1 || taps[0].Args[0] != 10 || taps[0].Args[1] != 20 || len(dev.CallsTo("PressKey")) != 1 {
		t.Fatalf("input not forwarded: %+v", dev.Calls())
	}

	status := testkit.Decode[lockResponse](t, e.call(t, "", http.MethodGet, "1/lock", nil))
	if !status.Locked || status.Lease == nil || status.Lease.Owner != "alice" || status.Lease.Token != "" {
		t.Fatalf("unexpected lock status %+v", status)
	}

	if rec := e.call(t, lease.Token, http.MethodDelete, "1/lock", nil); rec.Code != http.StatusNoContent {
		t.Fatalf("release: expected 204, got %d", rec.Code)
	}
	if dev.IsLocked() {
		t.Fatal("device still locked after release")
	}
}

func TestReleaseWithWrongTokenFails(t *testing.T) {
	e := setup(t)
	e.lock(t, "1", "alice")
	if rec := e.call(t, "bogus", http.MethodDelete, "1/lock", nil); rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d %s", rec.Code, rec.Body)
	}
}

// lockedByTest is a device that a test run holds: nobody else can lock it.
type lockedByTest struct{ *testkit.FakeDevice }

func (lockedByTest) Lock() error    { return errors.New("locked") }
func (lockedByTest) IsLocked() bool { return true }

func TestLockedByTestRun(t *testing.T) {
	e := setup(t)
	e.devices["1"] = lockedByTest{e.full}
	if rec := e.call(t, "", http.MethodPost, "1/lock", nil); rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	status := testkit.Decode[lockResponse](t, e.call(t, "", http.MethodGet, "1/lock", nil))
	if !status.LockedByTest || status.Lease != nil {
		t.Fatalf("expected lockedByTest, got %+v", status)
	}
}

func TestUnknownDeviceIsNotServed(t *testing.T) {
	e := setup(t)
	if rec := e.call(t, "", http.MethodGet, "9/lock", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
