package node

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type pushRecorder struct {
	mu     sync.Mutex
	pushes []DevicesChangedRequest
	tokens []string
	fail   int // answer this many requests with 503 first
}

func (r *pushRecorder) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/node/devices" || req.Method != http.MethodPost {
			t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.fail > 0 {
			r.fail--
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		var body DevicesChangedRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Errorf("decode push: %v", err)
		}
		r.pushes = append(r.pushes, body)
		r.tokens = append(r.tokens, req.Header.Get("X-Auth-Token"))
		w.WriteHeader(http.StatusAccepted)
	}
}

func (r *pushRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.pushes)
}

type deviceList struct {
	mu      sync.Mutex
	devices []*DeviceResponse
}

func (l *deviceList) set(devices ...*DeviceResponse) {
	l.mu.Lock()
	l.devices = devices
	l.mu.Unlock()
}

func (l *deviceList) snapshot() []*DeviceResponse {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]*DeviceResponse, len(l.devices))
	copy(out, l.devices)
	return out
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestDevicesNotifierPushesCoalescedFullList(t *testing.T) {
	rec := &pushRecorder{}
	srv := httptest.NewServer(rec.handler(t))
	defer srv.Close()

	list := &deviceList{}
	list.set(
		&DeviceResponse{Manager: "ios_device", DeviceID: "b", State: DeviceState_Booted},
		&DeviceResponse{Manager: "android_device", DeviceID: "a", State: DeviceState_Unauthorized},
	)
	token := "node-token"
	n := NewDevicesNotifier(strings.TrimPrefix(srv.URL, "http://"), "node-1", &token, list.snapshot)
	for i := 0; i < 5; i++ {
		n.Notify()
	}
	time.Sleep(DevicesNotifierDelay + 300*time.Millisecond)

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.pushes) != 1 {
		t.Fatalf("expected one coalesced push, got %d", len(rec.pushes))
	}
	p := rec.pushes[0]
	if p.Node != "node-1" || rec.tokens[0] != token || p.SentAt == 0 {
		t.Fatalf("unexpected push %+v token %q", p, rec.tokens[0])
	}
	if p.Devices == nil || len(*p.Devices) != 2 {
		t.Fatalf("expected the full device list, got %+v", p.Devices)
	}
	// sorted by manager and id, so equal lists compare equal
	if d := (*p.Devices)[0]; d.DeviceID != "a" || d.State != DeviceState_Unauthorized {
		t.Fatalf("unexpected first device %+v", d)
	}
}

func TestDevicesNotifierSkipsUnchangedList(t *testing.T) {
	rec := &pushRecorder{}
	srv := httptest.NewServer(rec.handler(t))
	defer srv.Close()

	list := &deviceList{}
	list.set(&DeviceResponse{Manager: "android_device", DeviceID: "a", State: DeviceState_Booted})
	n := NewDevicesNotifier(strings.TrimPrefix(srv.URL, "http://"), "node-1", nil, list.snapshot)

	n.Notify()
	waitFor(t, "first push", func() bool { return rec.count() == 1 })

	// a local refresh without changes doesn't reach the master
	n.Notify()
	time.Sleep(DevicesNotifierDelay + 200*time.Millisecond)
	if c := rec.count(); c != 1 {
		t.Fatalf("unchanged list was pushed again (%d pushes)", c)
	}

	list.set(&DeviceResponse{Manager: "android_device", DeviceID: "a", State: DeviceState_RemoteDisconnected})
	n.Notify()
	waitFor(t, "push of the change", func() bool { return rec.count() == 2 })

	// an empty list is a change too, and is sent as [] (not "no list")
	list.set()
	n.Notify()
	waitFor(t, "push of the empty list", func() bool { return rec.count() == 3 })
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if d := rec.pushes[2].Devices; d == nil || len(*d) != 0 {
		t.Fatalf("expected an empty device list, got %+v", d)
	}
}

func TestDevicesNotifierRetriesFailedPush(t *testing.T) {
	rec := &pushRecorder{fail: 1}
	srv := httptest.NewServer(rec.handler(t))
	defer srv.Close()

	list := &deviceList{}
	list.set(&DeviceResponse{Manager: "android_device", DeviceID: "a"})
	n := NewDevicesNotifier(strings.TrimPrefix(srv.URL, "http://"), "node-1", nil, list.snapshot)

	n.Notify()
	waitFor(t, "retried push", func() bool { return rec.count() == 1 })
}

func TestDevicesNotifierWithoutMasterIsNoop(t *testing.T) {
	n := NewDevicesNotifier("", "node-1", nil, func() []*DeviceResponse { return nil })
	if n != nil {
		t.Fatal("expected nil notifier without master")
	}
	n.Notify() // must not panic
}
