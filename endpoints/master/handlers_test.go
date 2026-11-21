package master

import (
	"net/http"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

type recordingHandler struct{ responses chan *appproto.Response }

func (h recordingHandler) OnActionResponse(_ interface{}, r *appproto.Response) { h.responses <- r }

func TestLog_DispatchesByType(t *testing.T) {
	tests := []struct {
		name   string
		body   map[string]any
		method string
		args   []any
	}{
		{"performance", map[string]any{"type": 0, "checkpoint": "c", "cpu": 1.0, "fps": 2.0, "mem": 3.0, "vertexCount": 4.0, "triangles": 5.0, "other": "o"}, "LogPerformance", []any{"c", 1.0, 2.0, 3.0, 4.0, 5.0, "o"}},
		{"data", map[string]any{"type": 1, "source": "s", "message": "m"}, "Data", []any{"s", "m"}},
		{"log", map[string]any{"type": 2, "source": "s", "message": "m"}, "Log", nil},
		{"error", map[string]any{"type": 3, "source": "s", "message": "m"}, "Error", nil},
		{"passed", map[string]any{"type": 4, "other": "passed"}, "Passed", []any{true}},
		{"failed", map[string]any{"type": 4, "other": "failed"}, "Passed", []any{false}},
		{"key", map[string]any{"type": 5, "key": "k", "value": "v"}, "SetKey", []any{"k", "v"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			dev := testkit.NewFakeDevice("d1")
			f.dm.Add("m", dev)
			tt.body["deviceId"] = "d1"
			rec := testkit.DoJSON(t, f.engine, http.MethodPost, "/node/log", tt.body)
			calls := dev.CallsTo(tt.method)
			if rec.Code != http.StatusOK || len(calls) != 1 {
				t.Fatalf("status %d, calls %v", rec.Code, dev.Calls())
			}
			for i, want := range tt.args {
				if calls[0].Args[i] != want {
					t.Errorf("arg %d = %v, want %v", i, calls[0].Args[i], want)
				}
			}
		})
	}
}

func TestLog_UnhandledTypeAndUnknownDeviceAreIgnored(t *testing.T) {
	f := newFixture(t)
	dev := testkit.NewFakeDevice("d1")
	f.dm.Add("m", dev)
	testkit.DoJSON(t, f.engine, http.MethodPost, "/node/log", map[string]any{"deviceId": "d1", "type": 99})
	testkit.DoJSON(t, f.engine, http.MethodPost, "/node/log", map[string]any{"deviceId": "nope", "type": 2})
	if len(dev.Calls()) != 0 {
		t.Errorf("unexpected calls %v", dev.Calls())
	}
}

func TestLog_BadBody(t *testing.T) {
	f := newFixture(t)
	rec := testkit.DoJSON(t, f.engine, http.MethodPost, "/node/log", "not an object")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestAction_ForwardsResponseToHandlers(t *testing.T) {
	f := newFixture(t)
	dev := testkit.NewFakeDevice("d1")
	h := recordingHandler{responses: make(chan *appproto.Response, 1)}
	dev.AddActionHandler(h)
	f.dm.Add("m", dev)
	rec := testkit.DoJSON(t, f.engine, http.MethodPost, "/node/action", map[string]any{"node": "n", "deviceId": "d1", "response": map[string]any{}})
	if rec.Code != http.StatusOK || len(h.responses) != 1 {
		t.Fatalf("status %d, responses %d", rec.Code, len(h.responses))
	}
}

func TestAction_UnknownDeviceAndBadBody(t *testing.T) {
	f := newFixture(t)
	if rec := testkit.DoJSON(t, f.engine, http.MethodPost, "/node/action", map[string]any{"deviceId": "nope"}); rec.Code != http.StatusOK {
		t.Errorf("unknown device status = %d", rec.Code)
	}
	if rec := testkit.DoJSON(t, f.engine, http.MethodPost, "/node/action", "bad"); rec.Code != http.StatusBadRequest {
		t.Errorf("bad body status = %d", rec.Code)
	}
}

func TestDevicesChanged(t *testing.T) {
	got := make(chan events.NodeDevicesChangedPayload, 4)
	events.NodeDevicesChanged.Register(events.HandlerFunc[events.NodeDevicesChangedPayload](func(p events.NodeDevicesChangedPayload) { got <- p }))
	f := newFixture(t)

	rec := testkit.DoJSON(t, f.engine, http.MethodPost, "/node/devices", map[string]any{"node": "n1", "sentAt": 5, "devices": []any{}})
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d", rec.Code)
	}
	select {
	case p := <-got:
		if p.NodeIdentifier != "n1" || p.SentAt != 5 {
			t.Errorf("payload = %+v", p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("event not triggered")
	}

	if rec := testkit.DoJSON(t, f.engine, http.MethodPost, "/node/devices", map[string]any{"node": "n1"}); rec.Code != http.StatusAccepted {
		t.Errorf("list-less push status = %d", rec.Code)
	}
	if rec := testkit.DoJSON(t, f.engine, http.MethodPost, "/node/devices", map[string]any{}); rec.Code != http.StatusBadRequest {
		t.Errorf("missing node status = %d", rec.Code)
	}
}
