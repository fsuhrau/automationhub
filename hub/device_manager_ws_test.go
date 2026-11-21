package hub

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// wsFixture serves the manager's /device/connect route; the websocket client
// plays the app.
type wsFixture struct {
	dm  *DeviceManager
	dev *testkit.FakeDevice
	srv *httptest.Server
}

func newWSFixture(t *testing.T) *wsFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dm := newTestManager(t)
	dev := testkit.NewFakeDevice("phone")
	dm.AddHandler(newFakeHandler("android_device", dev))
	r := gin.New()
	if err := dm.RegisterRoutes(r, r.Group("")); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return &wsFixture{dm: dm, dev: dev, srv: srv}
}

func (f *wsFixture) dial(t *testing.T) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(f.srv.URL, "http") + "/device/connect"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func handshake(deviceID string) appproto.Response {
	return appproto.Response{Payload: appproto.ResponseData{Connect: &appproto.Connect{DeviceID: deviceID, SessionID: "s1"}}}
}

func (f *wsFixture) connectApp(t *testing.T) (*websocket.Conn, *device.Connection) {
	t.Helper()
	conn := f.dial(t)
	if err := conn.WriteJSON(handshake("phone")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "device connection", func() bool { return f.dev.Connection() != nil })
	return conn, f.dev.Connection()
}

func TestAppConnectionLifecycle(t *testing.T) {
	f := newWSFixture(t)
	app, con := f.connectApp(t)
	if con.ConnectionParameter.SessionID != "s1" {
		t.Fatalf("connect parameter = %+v", con.ConnectionParameter)
	}

	// an action response reaches the waiting caller
	if err := app.WriteJSON(appproto.Response{ActionID: "a1", ActionType: appproto.ActionType_GetScreenshot}); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-con.ActionChannel:
		if got.ActionID != "a1" {
			t.Fatalf("response = %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("response not dispatched")
	}

	// a server action reaches the app
	f.dev.AppConnected = true
	if err := f.dev.Connection().Send([]byte(`{"x":1}`)); err != nil {
		t.Fatal(err)
	}
	if _, msg, err := app.ReadMessage(); err != nil || string(msg) != `{"x":1}` {
		t.Fatalf("app got %q, %v", msg, err)
	}

	// closing the app cleans the device up
	_ = app.Close()
	waitFor(t, "disconnect cleanup", func() bool { return f.dev.Connection() == nil })
}

func TestHandshakeForUnknownDeviceIsClosed(t *testing.T) {
	f := newWSFixture(t)
	conn := f.dial(t)
	if err := conn.WriteJSON(handshake("ghost")); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("server should close the connection")
	}
	if f.dev.Connection() != nil {
		t.Fatal("no device may be attached")
	}
}

func TestBrokenHandshakesAreRejected(t *testing.T) {
	tests := map[string]func(*websocket.Conn){
		"not json":        func(c *websocket.Conn) { _ = c.WriteMessage(websocket.TextMessage, []byte("{")) },
		"no connect data": func(c *websocket.Conn) { _ = c.WriteJSON(appproto.Response{}) },
		"closed at once":  func(c *websocket.Conn) { _ = c.Close() },
	}
	for name, send := range tests {
		t.Run(name, func(t *testing.T) {
			f := newWSFixture(t)
			conn := f.dial(t)
			send(conn)
			time.Sleep(50 * time.Millisecond)
			if f.dev.Connection() != nil {
				t.Fatal("no device may be attached")
			}
		})
	}
}

func TestConnectWithoutUpgradeFails(t *testing.T) {
	f := newWSFixture(t)
	resp, err := http.Get(f.srv.URL + "/device/connect")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestOriginCheck(t *testing.T) {
	tests := []struct {
		origin string
		cors   []string
		want   bool
	}{
		{"", nil, true},
		{"http://hub", []string{"http://hub"}, true},
		{"http://evil", []string{"http://hub"}, false},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if tt.origin != "" {
			req.Header.Set("Origin", tt.origin)
		}
		if got := newUpgrader(tt.cors).CheckOrigin(req); got != tt.want {
			t.Errorf("origin %q cors %v: got %v", tt.origin, tt.cors, got)
		}
	}
}

func TestHandleConnectionSurvivesAClosedSocket(t *testing.T) {
	f := newWSFixture(t)
	served := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := f.dm.upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		_ = conn.UnderlyingConn().Close() // deadline and read both fail
		f.dm.handleConnection(conn)
		close(served)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	<-served
}
