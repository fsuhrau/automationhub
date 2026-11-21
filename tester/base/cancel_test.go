package base

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gorilla/websocket"
)

// appDevice is a device whose app can be stopped; the rest of the device
// interface is not used by these tests
type appDevice struct {
	device.Device
	gen     generic.Device
	stopped bool
}

func (d *appDevice) Connection() *device.Connection     { return d.gen.Connection() }
func (d *appDevice) SetConnection(c *device.Connection) { d.gen.SetConnection(c) }
func (d *appDevice) IsAppConnected() bool               { return d.gen.Connection() != nil }
func (d *appDevice) StopApp(*app.Parameter) error {
	d.stopped = true
	return nil
}
func (d *appDevice) Log(string, string, ...interface{}) {}

// appSocket opens a websocket like an app's connection; the returned channel
// is closed once the server side noticed that the socket is closed
func appSocket(t *testing.T) (*websocket.Conn, <-chan struct{}) {
	t.Helper()
	closed := make(chan struct{})
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_, _, _ = conn.ReadMessage()
		close(closed)
	}))
	t.Cleanup(srv.Close)
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	return ws, closed
}

// Stopping the app forgets its connection and closes the socket. A killed
// app's socket can look open for minutes (through adb reverse), and the next
// run took it for its new app: the tests went into the dead socket.
func TestStopAndClearAppDropsTheConnection(t *testing.T) {
	ws, closed := appSocket(t)
	d := &appDevice{}
	d.SetConnection(&device.Connection{Connection: ws, ConnectionParameter: &action.Connect{DeviceID: "phone"}})

	if err := StopAndClearApp(d, &app.Parameter{Identifier: "com.example"}); err != nil {
		t.Fatal(err)
	}
	if !d.stopped || d.IsAppConnected() {
		t.Fatalf("stopped %v, still connected %v", d.stopped, d.IsAppConnected())
	}
	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		t.Fatal("the socket of the stopped app is still open")
	}
}

// A connection that replaced the dropped one stays.
func TestDropAppConnectionWithoutConnection(t *testing.T) {
	d := &appDevice{}
	device.DropAppConnection(d)
	newer := &device.Connection{ConnectionParameter: &action.Connect{}}
	d.SetConnection(newer)
	device.DropAppConnection(&appDevice{})
	if d.Connection() != newer {
		t.Fatal("another device's drop changed this one")
	}
}

// A cancel waits until the run released its devices, but not forever.
func TestWaitFinished(t *testing.T) {
	tr := NewRunner(Deps{}, models.TestTypeUnity, func(*Runner) Strategy { return nil })
	if tr.WaitFinished(10 * time.Millisecond) {
		t.Fatal("finished before the run ended")
	}
	go func() {
		time.Sleep(20 * time.Millisecond)
		tr.Finished()
	}()
	if !tr.WaitFinished(5 * time.Second) {
		t.Fatal("not finished after the run ended")
	}
	tr.Finished() // a second call is harmless
}
