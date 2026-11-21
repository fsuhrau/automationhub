package hub

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/rpc"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/hub/node/jsonrpc"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	_ manager.Nodes        = (*NodeManager)(nil)
	_ videostream.Provider = (*NodeManager)(nil)
)

// fakeRPC is the RPC connection of a node; it records every call.
type fakeRPC struct {
	mu       sync.Mutex
	calls    []string
	args     map[string][]any
	err      error
	uploaded bool
	pingErr  error
}

func (f *fakeRPC) rec(method string, args ...any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.args == nil {
		f.args = map[string][]any{}
	}
	f.calls = append(f.calls, method)
	f.args[method] = args
	return f.err
}

func (f *fakeRPC) called(method string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.args[method]
	return ok
}

func (f *fakeRPC) Ping() error { return f.pingErr }
func (f *fakeRPC) GetDevices() (map[string][]device.Device, error) {
	return map[string][]device.Device{"m": nil}, f.rec("GetDevices")
}
func (f *fakeRPC) StartDevice(id string) error { return f.rec("StartDevice", id) }
func (f *fakeRPC) StopDevice(id string) error  { return f.rec("StopDevice", id) }
func (f *fakeRPC) IsAppInstalled(id string, p *app.Parameter) (bool, error) {
	return true, f.rec("IsAppInstalled", id, p)
}
func (f *fakeRPC) InstallApp(id string, p *app.Parameter) error { return f.rec("InstallApp", id, p) }
func (f *fakeRPC) UninstallApp(id string, p *app.Parameter) error {
	return f.rec("UninstallApp", id, p)
}
func (f *fakeRPC) StartApp(id string, c *device.DeviceConfig, p *app.Parameter, s, u string) error {
	return f.rec("StartApp", id, c, p, s, u)
}
func (f *fakeRPC) StopApp(id string, p *app.Parameter) error { return f.rec("StopApp", id, p) }
func (f *fakeRPC) ClearAppData(id string, p *app.Parameter) error {
	return f.rec("ClearAppData", id, p)
}
func (f *fakeRPC) UploadApp(_ context.Context, p *app.Parameter) error {
	return f.rec("UploadApp", p)
}
func (f *fakeRPC) IsAppUploaded(p *app.Parameter) (bool, error) {
	return f.uploaded, f.rec("IsAppUploaded", p)
}
func (f *fakeRPC) Control(r device.ControlRequest) (device.ControlResponse, error) {
	return device.ControlResponse{Lines: []string{"x"}}, f.rec("Control", r)
}
func (f *fakeRPC) StartVideoStream(r device.VideoStreamRequest) error {
	return f.rec("StartVideoStream", r)
}
func (f *fakeRPC) IsConnected(id string) bool { _ = f.rec("IsConnected", id); return true }
func (f *fakeRPC) StartRecording(id, path string) error {
	return f.rec("StartRecording", id, path)
}
func (f *fakeRPC) StopRecording(id string) error { return f.rec("StopRecording", id) }
func (f *fakeRPC) GetScreenshot(id string) ([]byte, int, int, error) {
	return []byte("png"), 2, 3, f.rec("GetScreenshot", id)
}
func (f *fakeRPC) HasFeature(id, feature string) bool {
	_ = f.rec("HasFeature", id, feature)
	return true
}
func (f *fakeRPC) Execute(id, data string) { _ = f.rec("Execute", id, data) }
func (f *fakeRPC) ConnectionTimeout(id string) time.Duration {
	_ = f.rec("ConnectionTimeout", id)
	return time.Minute
}
func (f *fakeRPC) RunNativeScript(id string, s []byte) { _ = f.rec("RunNativeScript", id, s) }
func (f *fakeRPC) SendAction(id string, a []byte)      { _ = f.rec("SendAction", id, a) }

func newTestNodeManager(t *testing.T) (*NodeManager, *fakeRPC) {
	t.Helper()
	nm := NewNodeManager(logrus.New(), nil) // forwarding needs no database
	rpcClient := &fakeRPC{}
	nm.RegisterNode("n1", rpcClient, "host", "1.2.3.4", "linux", "env", 9, []string{"m"})
	return nm, rpcClient
}

// nodeCall is one forwarded call: what it does on a node and what it returns
// when the node can't be reached (the zero value or a fallback).
type nodeCall struct {
	method   string
	call     func(nm *NodeManager, id manager.NodeIdentifier) any
	wantArgs []any // arguments of the RPC call
	fallback any   // result without a node; an error is expected when it is the lookup error
}

func forwardedCalls() []nodeCall {
	p := &app.Parameter{Identifier: "com.demo"}
	cfg := &device.DeviceConfig{IP: "1.1.1.1"}
	req := device.ControlRequest{DeviceID: "d", Op: device.ControlTap}
	vreq := device.VideoStreamRequest{DeviceID: "d", StreamID: "s"}
	return []nodeCall{
		{"StartDevice", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.StartDevice(id, "d") }, []any{"d"}, nil},
		{"StopDevice", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.StopDevice(id, "d") }, []any{"d"}, nil},
		{"IsAppInstalled", func(nm *NodeManager, id manager.NodeIdentifier) any {
			ok, err := nm.IsAppInstalled(id, "d", p)
			return [2]any{ok, err}
		}, []any{"d", p}, nil},
		{"IsAppUploaded", func(nm *NodeManager, id manager.NodeIdentifier) any {
			ok, err := nm.IsAppUploaded(id, p)
			return [2]any{ok, err}
		}, []any{p}, nil},
		{"InstallApp", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.InstallApp(id, "d", p) }, []any{"d", p}, nil},
		{"UninstallApp", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.UninstallApp(id, "d", p) }, []any{"d", p}, nil},
		{"StartApp", func(nm *NodeManager, id manager.NodeIdentifier) any {
			return nm.StartApp(id, "d", cfg, p, "sess", "url")
		}, []any{"d", cfg, p, "sess", "url"}, nil},
		{"StopApp", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.StopApp(id, "d", p) }, []any{"d", p}, nil},
		{"ClearAppData", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.ClearAppData(id, "d", p) }, []any{"d", p}, nil},
		{"Control", func(nm *NodeManager, id manager.NodeIdentifier) any {
			r, err := nm.Control(id, req)
			return [2]any{r.Lines, err}
		}, []any{req}, nil},
		{"StartVideoStream", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.StartVideoStream(id, vreq) }, []any{vreq}, nil},
		{"IsConnected", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.IsConnected(id, "d") }, []any{"d"}, false},
		{"StartRecording", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.StartRecording(id, "d", "/ignored") }, []any{"d", ""}, nil},
		{"StopRecording", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.StopRecording(id, "d") }, []any{"d"}, nil},
		{"GetScreenshot", func(nm *NodeManager, id manager.NodeIdentifier) any {
			b, w, h, err := nm.GetScreenshot(id, "d")
			return [4]any{string(b), w, h, err}
		}, []any{"d"}, nil},
		{"HasFeature", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.HasFeature(id, "d", "f") }, []any{"d", "f"}, false},
		{"Execute", func(nm *NodeManager, id manager.NodeIdentifier) any { nm.Execute(id, "d", "x"); return nil }, []any{"d", "x"}, nil},
		{"ConnectionTimeout", func(nm *NodeManager, id manager.NodeIdentifier) any { return nm.ConnectionTimeout(id, "d") }, []any{"d"}, NodeConnectionTimeout},
		{"RunNativeScript", func(nm *NodeManager, id manager.NodeIdentifier) any {
			nm.RunNativeScript(id, "d", []byte("s"))
			return nil
		}, []any{"d", []byte("s")}, nil},
		{"SendAction", func(nm *NodeManager, id manager.NodeIdentifier) any { nm.SendAction(id, "d", []byte("a")); return nil }, []any{"d", []byte("a")}, nil},
	}
}

func TestForwarding_ReachesTheNodesConnection(t *testing.T) {
	for _, tc := range forwardedCalls() {
		t.Run(tc.method, func(t *testing.T) {
			nm, rpcClient := newTestNodeManager(t)
			tc.call(nm, "n1")
			if !rpcClient.called(tc.method) {
				t.Fatalf("%s did not reach the connection: %v", tc.method, rpcClient.calls)
			}
			if got := rpcClient.args[tc.method]; !reflect.DeepEqual(got, tc.wantArgs) {
				t.Fatalf("args = %v, want %v", got, tc.wantArgs)
			}
		})
	}
}

func TestForwarding_ReturnsTheNodesAnswerAndError(t *testing.T) {
	nm, rpcClient := newTestNodeManager(t)
	if got := nm.IsConnected("n1", "d"); !got {
		t.Error("IsConnected")
	}
	if got := nm.ConnectionTimeout("n1", "d"); got != time.Minute {
		t.Errorf("ConnectionTimeout = %v", got)
	}
	if ok, err := nm.IsAppInstalled("n1", "d", nil); !ok || err != nil {
		t.Errorf("IsAppInstalled = %v, %v", ok, err)
	}
	if b, w, h, err := nm.GetScreenshot("n1", "d"); string(b) != "png" || w != 2 || h != 3 || err != nil {
		t.Errorf("GetScreenshot = %q %d %d %v", b, w, h, err)
	}
	boom := errors.New("node error")
	rpcClient.err = boom
	if err := nm.StartDevice("n1", "d"); !errors.Is(err, boom) {
		t.Errorf("StartDevice = %v", err)
	}
	if _, err := nm.Control("n1", device.ControlRequest{}); !errors.Is(err, boom) {
		t.Errorf("Control = %v", err)
	}
	if _, _, _, err := nm.GetScreenshot("n1", "d"); !errors.Is(err, boom) {
		t.Errorf("GetScreenshot = %v", err)
	}
}

func TestForwarding_UnknownNode(t *testing.T) {
	for _, tc := range forwardedCalls() {
		t.Run(tc.method, func(t *testing.T) {
			nm, _ := newTestNodeManager(t)
			got := tc.call(nm, "unknown")
			switch want := tc.fallback; {
			case want != nil:
				if got != want {
					t.Errorf("fallback = %v, want %v", got, want)
				}
			case got == nil:
				// fire-and-forget calls are dropped
			default:
				assertLookupError(t, got, ErrNodeNotFound)
			}
		})
	}
}

// assertLookupError checks that the result is, or ends with, ErrNodeNotFound.
func assertLookupError(t *testing.T, got any, want error) {
	t.Helper()
	switch v := got.(type) {
	case error:
		if !errors.Is(v, want) {
			t.Errorf("err = %v", v)
		}
	case [2]any:
		assertLookupError(t, v[1], want)
	case [4]any:
		assertLookupError(t, v[3], want)
	default:
		t.Errorf("unexpected result %#v", got)
	}
}

func TestForwarding_DisconnectedNode(t *testing.T) {
	nm, _ := newTestNodeManager(t)
	nm.mu.Lock()
	n := nm.nodes["n1"]
	n.Handler, n.Status = nil, NodeStatusDisconnected
	nm.nodes["n1"] = n
	nm.mu.Unlock()
	if err := nm.StartDevice("n1", "d"); !errors.Is(err, ErrNodeNotConnected) {
		t.Errorf("err = %v", err)
	}
	if nm.IsConnected("n1", "d") {
		t.Error("a disconnected node has no connected device")
	}
}

func TestUploadApp_SkipsWhatTheNodeHas(t *testing.T) {
	p := &app.Parameter{}
	nm, rpcClient := newTestNodeManager(t)
	if err := nm.UploadApp(context.Background(), "n1", p); err != nil || !rpcClient.called("UploadApp") {
		t.Fatalf("a missing app is uploaded: %v, %v", err, rpcClient.calls)
	}

	nm, rpcClient = newTestNodeManager(t)
	rpcClient.uploaded = true
	if err := nm.UploadApp(context.Background(), "n1", p); err != nil || rpcClient.called("UploadApp") {
		t.Fatalf("an uploaded app is not uploaded again: %v, %v", err, rpcClient.calls)
	}

	nm, rpcClient = newTestNodeManager(t)
	rpcClient.err = errors.New("check failed")
	if err := nm.UploadApp(context.Background(), "n1", p); err == nil || rpcClient.called("UploadApp") {
		t.Fatalf("a failed check stops the upload: %v", err)
	}
	if err := nm.UploadApp(context.Background(), "unknown", p); !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestNodeManager_Rendezvous(t *testing.T) {
	nm := NewNodeManager(logrus.New(), nil)
	if nm.Rendezvous() == nil || videostream.From(nm) != nm.Rendezvous() {
		t.Error("the node manager owns one rendezvous")
	}
	if NewNodeManager(logrus.New(), nil).Rendezvous() == nm.Rendezvous() {
		t.Error("each node manager has its own")
	}
}

func TestNodeManager_Registry(t *testing.T) {
	nm, _ := newTestNodeManager(t)
	if got := nm.GetNodes(); len(got) != 1 || got[0] != "n1" {
		t.Errorf("GetNodes = %v", got)
	}
	if status, err := nm.GetStatus("n1"); err != nil || status != int(NodeStatusConnected) {
		t.Errorf("GetStatus = %d, %v", status, err)
	}
	if _, err := nm.GetStatus("x"); !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("GetStatus(unknown) = %v", err)
	}
	if got, err := nm.GetManagers("n1"); err != nil || len(got) != 1 {
		t.Errorf("GetManagers = %v, %v", got, err)
	}
	if _, err := nm.GetManagers("x"); err == nil {
		t.Error("GetManagers(unknown)")
	}
	withDB := NewNodeManager(logrus.New(), testkit.DB(t))
	if withDB.IsNodeKnown("n1") {
		t.Error("only nodes in the database are known")
	}
}

func TestNodeManager_HealthCheckDisconnectsUnreachableNodes(t *testing.T) {
	nm, rpcClient := newTestNodeManager(t)
	lost := make(chan manager.NodeIdentifier, 1)
	events.NodeDisconnected.Register(disconnectedFunc(func(p events.NodeDisconnectedPayload) { lost <- p.NodeIdentifier }))

	nm.healthCheck()
	if status, _ := nm.GetStatus("n1"); status != int(NodeStatusConnected) {
		t.Fatal("a node that answers stays connected")
	}
	rpcClient.pingErr = errors.New("no pong")
	nm.healthCheck()
	if status, _ := nm.GetStatus("n1"); status != int(NodeStatusDisconnected) {
		t.Fatal("a node that does not answer is disconnected")
	}
	if got := <-lost; got != "n1" {
		t.Errorf("event for %q", got)
	}
}

type disconnectedFunc func(events.NodeDisconnectedPayload)

func (f disconnectedFunc) Handle(p events.NodeDisconnectedPayload) { f(p) }

func TestNodeManager_RunStopsWithTheContext(t *testing.T) {
	nm, _ := newTestNodeManager(t)
	ctx, cancel := context.WithCancel(context.Background())
	nm.Run(ctx)
	cancel()
}

// --- the whole chain over a real websocket ---

// TestNodeManager_ForwardsToARealNode drives a device of a node through the
// node manager, the RPC client, the jsonrpc codec and the node's RPC receiver.
func TestNodeManager_ForwardsToARealNode(t *testing.T) {
	dm := &testkit.FakeDeviceManager{}
	dev := testkit.NewFakeDevice("phone")
	dev.AppConnected, dev.Timeout, dev.Features = true, 7*time.Second, []string{"lua"}
	dev.PNG, dev.Width, dev.Height = []byte("png"), 4, 5
	dm.Add("android_device", dev)

	rpcNode, err := node.NewRPCNode(config.Service{Paths: config.Paths{Bundles: t.TempDir()}}, dm, nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		server := rpc.NewServer()
		_ = server.RegisterName("RPCNode", rpcNode)
		server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}))
	defer srv.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	nm := NewNodeManager(logrus.New(), testkit.DB(t))
	nm.RegisterNode("n1", node.NewRPCClient(conn, "master"), "host", "addr", "linux", "env", 1, nil)

	if !nm.IsConnected("n1", "phone") || !nm.HasFeature("n1", "phone", "lua") || nm.ConnectionTimeout("n1", "phone") != 7*time.Second {
		t.Error("device queries did not reach the device")
	}
	if b, w, h, err := nm.GetScreenshot("n1", "phone"); err != nil || string(b) != "png" || w != 4 || h != 5 {
		t.Errorf("screenshot = %q %d %d %v", b, w, h, err)
	}
	if err := nm.StartDevice("n1", "phone"); err != nil || len(dm.CallsTo("Start")) != 1 {
		t.Errorf("StartDevice: %v", err)
	}
	if err := nm.StartDevice("n1", "ghost"); err == nil || err.Error() != "device with id ghost not found" {
		t.Errorf("StartDevice(ghost) = %v", err)
	}
	nm.Execute("n1", "phone", "cmd")
	if len(dev.CallsTo("Execute")) != 1 {
		t.Error("Execute did not reach the device")
	}
	if got, err := nm.GetManagers("n1"); err != nil || len(got["android_device"]) != 1 {
		t.Errorf("GetManagers = %v, %v", got, err)
	}
}
