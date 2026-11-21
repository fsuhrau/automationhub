package node

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/rpc"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/node/jsonrpc"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gorilla/websocket"
)

// wireRecorder wraps the node's server codec and keeps the raw params of
// every request, so the tests pin the exact JSON the master sends.
type wireRecorder struct {
	rpc.ServerCodec
	host string

	mu     sync.Mutex
	method string
	lines  []string
}

func (w *wireRecorder) ReadRequestHeader(r *rpc.Request) error {
	if err := w.ServerCodec.ReadRequestHeader(r); err != nil {
		return err
	}
	w.mu.Lock()
	w.method = r.ServiceMethod
	w.mu.Unlock()
	return nil
}

func (w *wireRecorder) ReadRequestBody(x any) error {
	if x == nil {
		return w.ServerCodec.ReadRequestBody(x)
	}
	var raw json.RawMessage
	if err := w.ServerCodec.ReadRequestBody(&raw); err != nil {
		return err
	}
	w.mu.Lock()
	w.lines = append(w.lines, w.method+" "+strings.ReplaceAll(string(raw), w.host, "master.test"))
	w.mu.Unlock()
	return json.Unmarshal(raw, x)
}

func (w *wireRecorder) golden() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return strings.Join(w.lines, "\n") + "\n"
}

// wire is a master RPCClient connected to a real RPCNode over a websocket.
type wire struct {
	client *RPCClient
	node   *RPCNode
	rec    *wireRecorder
	dev    *testkit.FakeDevice
	dm     *testkit.FakeDeviceManager
	bundle *AppBundleManager
	// videoPaths receives the path of every stream the node dials back
	videoPaths chan string
}

func newWire(t *testing.T) *wire {
	t.Helper()
	return newWireFor(t, nil)
}

// newWireFor is newWire with receiver served as "RPCNode" instead of the real
// RPCNode; nil keeps the real one.
func newWireFor(t *testing.T, receiver any) *wire {
	t.Helper()
	w := &wire{dev: testkit.NewFakeDevice("dev-1"), dm: &testkit.FakeDeviceManager{}, videoPaths: make(chan string, 4)}
	w.dm.Add("android_device", w.dev)
	var err error
	if w.bundle, err = NewAppBundleManager(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if err := w.bundle.StoreData([]byte("data"), &AppBundleMetaData{Filename: "app.apk", FileHash: "h1", FileSize: 4}); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")
	w.node = &RPCNode{
		config:         config.Service{MasterURL: host, NodeUrl: "node.test"},
		dm:             w.dm,
		abm:            w.bundle,
		uploadRequests: map[int32]*UploadProgress{},
		screenshots:    map[string][]byte{},
	}
	w.rec = &wireRecorder{host: host}
	mux.HandleFunc("/rpc", func(rw http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(rw, r, nil)
		if err != nil {
			return
		}
		server := rpc.NewServer()
		if receiver == nil {
			receiver = w.node
		}
		_ = server.RegisterName("RPCNode", receiver)
		w.rec.ServerCodec = jsonrpc.NewServerCodec(conn)
		server.ServeCodec(w.rec)
	})
	mux.HandleFunc("/upload/", func(rw http.ResponseWriter, r *http.Request) { _, _ = rw.Write([]byte("data")) })
	mux.HandleFunc("/node/video/", func(rw http.ResponseWriter, r *http.Request) {
		w.videoPaths <- r.URL.Path
		if conn, err := (&websocket.Upgrader{}).Upgrade(rw, r, nil); err == nil {
			_ = conn.Close()
		}
	})

	conn, _, err := websocket.DefaultDialer.Dial("ws://"+host+"/rpc", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	w.client = NewRPCClient(conn, host)
	w.client.uploadInterval = 10 * time.Millisecond
	return w
}

func testDeviceNamed(id string) *testkit.FakeDevice { return testkit.NewFakeDevice(id) }

func netrpcServerError(msg string) error { return rpc.ServerError(msg) }
