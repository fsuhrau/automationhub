package manager

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func newServer(t *testing.T) (*httptest.Server, *testkit.FakeDeviceManager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	dm := &testkit.FakeDeviceManager{}
	r := gin.New()
	if err := New(logger, dm, config.Service{}).RegisterRoutes(r, r.Group("/")); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(r)
	t.Cleanup(ts.Close)
	return ts, dm
}

func dial(t *testing.T, ts *httptest.Server) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http")+"/manager/register", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func waitForCalls(t *testing.T, dm *testkit.FakeDeviceManager, n int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for len(dm.CallsTo("RegisterDevice")) != n {
		if time.Now().After(deadline) {
			t.Fatalf("RegisterDevice calls = %d, want %d", len(dm.CallsTo("RegisterDevice")), n)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestRegister_PassesDeviceAndConnectionToManager(t *testing.T) {
	ts, dm := newServer(t)
	conn := dial(t, ts)
	msg := `{"Type":"android","DeviceID":"d1","Version":"14","OS":"android","TargetVersion":"34","Name":"Pixel","DeviceModel":"P8","RAM":8,"GPU":"g","SOC":"s","ProjectDir":"/p"}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatal(err)
	}
	waitForCalls(t, dm, 1)
	got := dm.CallsTo("RegisterDevice")[0].Args[0].(device.RegisterData)
	if got.DeviceID != "d1" || got.ManagerType != "android" || got.Name != "Pixel" || got.RAM != 8 ||
		got.DeviceOSVersion != "14" || got.TargetVersion != "34" || got.ProjectDir != "/p" ||
		got.DeviceIP == nil || got.Conn == nil {
		t.Errorf("register data = %+v", got)
	}
}

func TestRegister_IgnoresBadMessages(t *testing.T) {
	ts, dm := newServer(t)
	_ = dial(t, ts).WriteMessage(websocket.TextMessage, []byte("not json"))
	_ = dial(t, ts).Close() // closes before sending a register request
	good := dial(t, ts)
	_ = good.WriteMessage(websocket.TextMessage, []byte(`{"DeviceID":"ok"}`))
	waitForCalls(t, dm, 1) // handlers run in order of arrival; only the valid one registers
	if id := dm.CallsTo("RegisterDevice")[0].Args[0].(device.RegisterData).DeviceID; id != "ok" {
		t.Errorf("registered %q", id)
	}
}

func TestRegister_UpgradeFailureReturnsBadRequest(t *testing.T) {
	ts, dm := newServer(t)
	resp, err := http.Get(ts.URL + "/manager/register")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest || len(dm.Calls()) != 0 {
		t.Errorf("status = %d, calls = %v", resp.StatusCode, dm.Calls())
	}
}
