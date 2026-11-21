package master

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func TestConnect_RegistersNode(t *testing.T) {
	f := newFixture(t)
	conn, _, err := f.dial("/node/connect")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	data, _ := proto.Marshal(&node.RegisterNodeRequest{Identifier: "n1", Hostname: "host", Port: 9, Managers: []string{"ios"}})
	if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		t.Fatal(err)
	}
	eventually(t, "node registration", func() bool { return len(f.nm.CallsTo("RegisterNode")) == 1 })
	args := f.nm.CallsTo("RegisterNode")[0].Args
	if args[2] != "host" || args[6] != int32(9) {
		t.Errorf("RegisterNode args = %v", args)
	}
}

func TestConnect_IgnoresBrokenRegistration(t *testing.T) {
	f := newFixture(t)
	conn, _, err := f.dial("/node/connect")
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.WriteMessage(websocket.BinaryMessage, []byte{0xff, 0xff, 0xff})
	_ = conn.Close()
	conn2, _, err := f.dial("/node/connect") // closed before sending anything
	if err != nil {
		t.Fatal(err)
	}
	_ = conn2.Close()
	if n := len(f.nm.CallsTo("RegisterNode")); n != 0 {
		t.Errorf("RegisterNode called %d times", n)
	}
}

func TestUpgradeFailure_ReturnsBadRequest(t *testing.T) {
	f := newFixture(t)
	for _, path := range []string{"/node/connect", "/node/video/s1"} {
		rec := httptest.NewRecorder()
		f.engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d", path, rec.Code)
		}
	}
}

// rendezvousNodes is a node manager that owns a rendezvous, like the real one.
type rendezvousNodes struct {
	*testkit.FakeNodes
	rendezvous *videostream.Rendezvous
}

func (n rendezvousNodes) Rendezvous() *videostream.Rendezvous { return n.rendezvous }

func TestVideo_DeliversConnectionToWaiter(t *testing.T) {
	rendezvous := videostream.NewRendezvous()
	engine := gin.New()
	nodes := rendezvousNodes{&testkit.FakeNodes{}, rendezvous}
	if err := New(config.Service{}, &testkit.FakeDeviceManager{}, nodes, nil).RegisterRoutes(engine, engine.Group("/")); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(engine)
	defer server.Close()

	ch, done := rendezvous.Expect("g35-delivered")
	defer done()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/node/video/g35-delivered", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if got := <-ch; got == nil {
		t.Fatal("no connection delivered")
	}
}

func TestVideo_ClosesConnectionWithoutWaiter(t *testing.T) {
	f := newFixture(t)
	conn, _, err := f.dial("/node/video/g35-nobody")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("expected the master to close the socket")
	}
}
