package hub

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

// recordingHandler is an appproto.ActionHandler that keeps what it received.
type recordingHandler struct {
	mu   sync.Mutex
	resp []*appproto.Response
}

func (h *recordingHandler) OnActionResponse(_ interface{}, r *appproto.Response) {
	h.mu.Lock()
	h.resp = append(h.resp, r)
	h.mu.Unlock()
}

func (h *recordingHandler) received() []*appproto.Response {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*appproto.Response(nil), h.resp...)
}

func newConnection() *device.Connection {
	return &device.Connection{
		ResponseChannel: make(chan device.ResponseData, 10),
		ActionChannel:   make(chan appproto.Response, 1),
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestDispatchRoutesEveryResponseKind(t *testing.T) {
	tests := []struct {
		name  string
		resp  appproto.Response
		check func(t *testing.T, d *testkit.FakeDevice, con *device.Connection)
	}{
		{"exception log", logResponse(appproto.LogLevel_Exception), expectCall("Exception")},
		{"error log", logResponse(appproto.LogLevel_Error), expectCall("Error")},
		{"info log", logResponse(appproto.LogLevel_Info), expectCall("Log")},
		{"log without data", appproto.Response{ActionType: appproto.ActionType_Log}, expectNoCall("Log")},
		{"run key", appproto.Response{ActionType: appproto.ActionType_RunKey, Payload: appproto.ResponseData{RunKey: &appproto.RunKey{Key: "k", Value: "v"}}}, expectCall("SetKey")},
		{"empty run key", appproto.Response{ActionType: appproto.ActionType_RunKey, Payload: appproto.ResponseData{RunKey: &appproto.RunKey{}}}, expectNoCall("SetKey")},
		{"run key without payload", appproto.Response{ActionType: appproto.ActionType_RunKey}, expectNoCall("SetKey")},
		{"performance dropped", appproto.Response{ActionType: appproto.ActionType_Performance, ActionID: "p"}, expectNoQueued},
		{"response with id queued", appproto.Response{ActionType: appproto.ActionType_GetScreenshot, ActionID: "a1"}, expectQueued("a1")},
		{"response without id logged", appproto.Response{ActionType: appproto.ActionType_GetScreenshot}, expectCall("Log")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := newTestManager(t)
			d := testkit.NewFakeDevice("phone")
			con := newConnection()
			dm.dispatchData(d, con, mustJSON(t, tt.resp))
			tt.check(t, d, con)
		})
	}
}

func logResponse(level appproto.LogLevel) appproto.Response {
	return appproto.Response{ActionType: appproto.ActionType_Log, Payload: appproto.ResponseData{LogData: &appproto.LogData{Level: level, Message: "m"}}}
}

func expectCall(method string) func(*testing.T, *testkit.FakeDevice, *device.Connection) {
	return func(t *testing.T, d *testkit.FakeDevice, _ *device.Connection) {
		t.Helper()
		if len(d.CallsTo(method)) != 1 {
			t.Fatalf("calls to %s = %v", method, d.Calls())
		}
	}
}

func expectNoCall(method string) func(*testing.T, *testkit.FakeDevice, *device.Connection) {
	return func(t *testing.T, d *testkit.FakeDevice, _ *device.Connection) {
		t.Helper()
		if len(d.CallsTo(method)) != 0 {
			t.Fatalf("unexpected call to %s", method)
		}
	}
}

func expectQueued(id string) func(*testing.T, *testkit.FakeDevice, *device.Connection) {
	return func(t *testing.T, _ *testkit.FakeDevice, con *device.Connection) {
		t.Helper()
		select {
		case got := <-con.ActionChannel:
			if got.ActionID != id {
				t.Fatalf("queued %q, want %q", got.ActionID, id)
			}
		default:
			t.Fatal("response not queued")
		}
	}
}

func expectNoQueued(t *testing.T, _ *testkit.FakeDevice, con *device.Connection) {
	t.Helper()
	if len(con.ActionChannel) != 0 {
		t.Fatal("response must not be queued")
	}
}

func TestDispatchPrefersRegisteredActionHandlers(t *testing.T) {
	dm := newTestManager(t)
	d := testkit.NewFakeDevice("phone")
	h1, h2 := &recordingHandler{}, &recordingHandler{}
	d.AddActionHandler(h1)
	d.AddActionHandler(h2)
	con := newConnection()

	dm.dispatchData(d, con, mustJSON(t, appproto.Response{ActionType: appproto.ActionType_GetScreenshot, ActionID: "a"}))

	if len(h1.received()) != 1 || len(h2.received()) != 1 {
		t.Fatal("every handler must see the response")
	}
	expectNoQueued(t, d, con)
}

func TestDispatchDropsResponseWhenNobodyListens(t *testing.T) {
	dm := newTestManager(t)
	d := testkit.NewFakeDevice("phone")
	con := newConnection()
	for _, id := range []string{"first", "second"} { // the channel holds one
		dm.dispatchData(d, con, mustJSON(t, appproto.Response{ActionType: appproto.ActionType_GetScreenshot, ActionID: id}))
	}
	expectQueued("first")(t, d, con)
	expectNoQueued(t, d, con)
}

func TestDispatchIgnoresInvalidJSON(t *testing.T) {
	dm := newTestManager(t)
	d := testkit.NewFakeDevice("phone")
	con := newConnection()
	dm.dispatchData(d, con, []byte("{"))
	expectNoQueued(t, d, con)
	if len(d.Calls()) != 0 {
		t.Fatalf("calls = %v", d.Calls())
	}
}

func TestHandleActionsWithoutConnectionReturns(t *testing.T) {
	dm := newTestManager(t)
	dm.handleActions(testkit.NewFakeDevice("phone"), context.Background())
}

func runHandleActions(dm *DeviceManager, d *testkit.FakeDevice, ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		dm.handleActions(d, ctx)
		close(done)
	}()
	return done
}

func TestHandleActionsEndsOnConnectionError(t *testing.T) {
	dm := newTestManager(t)
	d := testkit.NewFakeDevice("phone")
	con := newConnection()
	d.SetConnection(con)
	done := runHandleActions(dm, d, context.Background())

	con.ResponseChannel <- device.ResponseData{Err: errors.New("closed")}
	<-done
	if d.Connection() != nil {
		t.Fatal("connection must be cleared")
	}
}

func TestHandleActionsEndsWhenChannelCloses(t *testing.T) {
	dm := newTestManager(t)
	d := testkit.NewFakeDevice("phone")
	con := newConnection()
	d.SetConnection(con)
	done := runHandleActions(dm, d, context.Background())
	close(con.ResponseChannel)
	<-done
}

func TestHandleActionsKeepsNewerConnection(t *testing.T) {
	dm := newTestManager(t)
	d := testkit.NewFakeDevice("phone")
	old, newer := newConnection(), newConnection()
	d.SetConnection(old)
	done := runHandleActions(dm, d, context.Background())
	// a handled response proves the loop runs on the old connection
	old.ResponseChannel <- device.ResponseData{Data: mustJSON(t, logResponse(appproto.LogLevel_Info))}
	waitFor(t, "first response", func() bool { return len(d.CallsTo("Log")) == 1 })
	d.SetConnection(newer)
	old.ResponseChannel <- device.ResponseData{Err: errors.New("closed")}
	<-done
	if d.Connection() != newer {
		t.Fatal("a newer connection must survive the old one's cleanup")
	}
}

func TestHandleActionsCancelAbortsExecutors(t *testing.T) {
	dm := newTestManager(t)
	d := testkit.NewFakeDevice("phone")
	h := &recordingHandler{}
	d.AddActionHandler(h)
	d.SetConnection(newConnection())
	ctx, cancel := context.WithCancel(context.Background())
	done := runHandleActions(dm, d, ctx)
	cancel()
	<-done
	got := h.received()
	if len(got) != 1 || got[0] != nil {
		t.Fatalf("handler got %v, want one nil response", got)
	}
}

func TestSendActionOutcomes(t *testing.T) {
	tests := []struct {
		name      string
		act       *fakeAction
		connected bool
		sendErr   error
		wantErr   bool
		wantSent  bool
	}{
		{"sent", &fakeAction{}, true, nil, false, true},
		{"marshal error", &fakeAction{err: errors.New("bad")}, true, nil, true, false},
		{"not connected", &fakeAction{}, false, nil, true, false},
		{"send error", &fakeAction{}, true, errors.New("write"), true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := newTestManager(t)
			d := testkit.NewFakeDevice("phone")
			d.AppConnected = tt.connected
			d.FailWith("Send", tt.sendErr)
			err := dm.SendAction(d, tt.act)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v", err)
			}
			if sent := len(d.CallsTo("Send")) == 1; sent != tt.wantSent {
				t.Fatalf("sent = %v, want %v", sent, tt.wantSent)
			}
		})
	}
}

func TestGetTypeOfLog(t *testing.T) {
	tests := map[appproto.LogType]string{
		appproto.LogType_DeviceLog:      "app",
		appproto.LogType_StepLog:        "step",
		appproto.LogType_StatusLog:      "status",
		appproto.LogType_CheckpointLog:  "checkpoint",
		appproto.LogType_PerformanceLog: "performance",
		appproto.LogType(99):            "",
	}
	for in, want := range tests {
		if got := getTypeOfLog(in); got != want {
			t.Errorf("getTypeOfLog(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestLogErrorReportsExceptions(t *testing.T) {
	d := testkit.NewFakeDevice("phone")
	if !logError(d, &appproto.LogData{Level: appproto.LogLevel_Exception}) {
		t.Fatal("exception must be reported")
	}
	if logError(d, &appproto.LogData{Level: appproto.LogLevel_Error}) {
		t.Fatal("error is not an exception")
	}
}
