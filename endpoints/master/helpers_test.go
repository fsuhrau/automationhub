package master

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type fixture struct {
	server *httptest.Server
	engine *gin.Engine
	dm     *testkit.FakeDeviceManager
	nm     *testkit.FakeNodes
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	f := &fixture{engine: gin.New(), dm: &testkit.FakeDeviceManager{}, nm: &testkit.FakeNodes{}}
	if err := New(config.Service{}, f.dm, f.nm, nil).RegisterRoutes(f.engine, f.engine.Group("/")); err != nil {
		t.Fatal(err)
	}
	f.server = httptest.NewServer(f.engine)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fixture) dial(path string) (*websocket.Conn, *http.Response, error) {
	return websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(f.server.URL, "http")+path, nil)
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
