package node

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/hub/node/jsonrpc"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

func newService(cfg config.Service) *Service {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return New(logger, nil, &node.RPCNode{}, nil, cfg, []string{"android"}, nil)
}

// fakeMaster accepts node connections and hands each one to onConn.
func fakeMaster(t *testing.T, onConn func(*websocket.Conn, http.Header)) string {
	t.Helper()
	upgrader := websocket.Upgrader{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		onConn(conn, r.Header)
	}))
	t.Cleanup(ts.Close)
	return strings.TrimPrefix(ts.URL, "http://")
}

func TestRegisterRoutes_TwoServicesDoNotClash(t *testing.T) {
	for i := 0; i < 2; i++ {
		if err := newService(config.Service{}).RegisterRoutes(nil, nil); err != nil {
			t.Fatalf("service %d: %v", i, err)
		}
	}
}

func TestRegisterRoutes_RegisteringTwiceOnOneServiceFails(t *testing.T) {
	s := newService(config.Service{})
	if err := s.RegisterRoutes(nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.RegisterRoutes(nil, nil); err == nil {
		t.Fatal("expected an error for the duplicate registration")
	}
}

type masterResult struct {
	registration node.RegisterNodeRequest
	token        string
	pingErr      error
}

func TestConnectAndServe_RegistersAndServesRPC(t *testing.T) {
	result := make(chan *masterResult, 1)
	host := fakeMaster(t, func(conn *websocket.Conn, h http.Header) {
		res := &masterResult{}
		res.token = h.Get("X-Auth-Token")
		_, data, err := conn.ReadMessage()
		if err == nil {
			err = proto.Unmarshal(data, &res.registration)
		}
		if err == nil {
			client := jsonrpc.NewClient(conn)
			res.pingErr = client.Call("RPCNode.Ping", &node.Void{}, &node.Void{})
			_ = client.Close()
		} else {
			res.pingErr = err
		}
		result <- res
	})
	cfg := config.Service{MasterURL: host, Identifier: "n1", Port: 7}
	cfg.Auth.Token = &config.Token{AuthToken: "secret"}
	s := newService(cfg)
	if err := s.RegisterRoutes(nil, nil); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() { s.ConnectAndServe(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ConnectAndServe did not return after the master closed")
	}

	res := <-result
	if res.pingErr != nil {
		t.Fatalf("ping: %v", res.pingErr)
	}
	if res.token != "secret" {
		t.Errorf("token header = %q", res.token)
	}
	if res.registration.GetIdentifier() != "n1" || res.registration.GetPort() != 7 || res.registration.GetManagers()[0] != "android" {
		t.Errorf("registration = %v", &res.registration)
	}
	if s.c != nil || s.dialing.Load() {
		t.Error("connection state not reset after serving")
	}
}

func TestConnectAndServe_DialFailureReleasesDialing(t *testing.T) {
	s := newService(config.Service{MasterURL: "127.0.0.1:1"})
	s.ConnectAndServe()
	if s.dialing.Load() || s.c != nil {
		t.Error("state not reset after a failed dial")
	}
}

func TestConnectAndServe_RegisterFailureClosesConnection(t *testing.T) {
	closed := make(chan struct{})
	host := fakeMaster(t, func(conn *websocket.Conn, _ http.Header) {
		_, _, _ = conn.ReadMessage() // returns when the node closes the socket
		close(closed)
	})
	s := newService(config.Service{MasterURL: host, Identifier: "\xff"}) // invalid UTF-8 cannot be marshaled
	s.ConnectAndServe()
	<-closed
	if s.dialing.Load() || s.c != nil {
		t.Error("state not reset after a failed registration")
	}
}

func TestConnectAndServe_SecondDialIsIgnored(t *testing.T) {
	s := newService(config.Service{MasterURL: "127.0.0.1:1"})
	s.dialing.Store(true)
	s.ConnectAndServe()
	if !s.dialing.Load() {
		t.Error("a running dial must keep its flag")
	}
}

func TestCloseConnection_ClosesActiveSocket(t *testing.T) {
	host := fakeMaster(t, func(conn *websocket.Conn, _ http.Header) { _, _, _ = conn.ReadMessage() })
	c, _, err := websocket.DefaultDialer.Dial("ws://"+host, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := newService(config.Service{})
	s.setConn(c)
	s.closeConnection()
	if s.c != nil {
		t.Fatal("connection not cleared")
	}
	if err := c.WriteMessage(websocket.TextMessage, nil); err == nil {
		t.Fatal("socket still open")
	}
}

func TestClearConn_KeepsNewerConnection(t *testing.T) {
	s := newService(config.Service{})
	older, newer := &websocket.Conn{}, &websocket.Conn{}
	s.setConn(newer)
	s.clearConn(older)
	if s.c != newer {
		t.Error("a newer connection must survive clearing an older one")
	}
}
