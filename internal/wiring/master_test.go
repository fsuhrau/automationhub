package wiring_test

import (
	"bufio"
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/internal/wiring"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// masterEnv is a wired master on an in-memory database and a fake device
// handler, served on an httptest server; the client carries the session
// cookie of a registered user.
type masterEnv struct {
	master *wiring.Master
	server *httptest.Server
	client *http.Client
}

func newMasterEnv(t *testing.T) *masterEnv {
	t.Helper()
	cfg := testConfig(t)
	cfg.Auth.Password = &config.Password{Secret: "test-secret-test-secret-test-sec"}
	db := testkit.DB(t)
	testkit.Project(db, func(p *models.Project) { p.Identifier, p.Name = "proj", "Project" })
	testkit.Device(db, func(d *models.Device) {
		d.DeviceIdentifier, d.Name, d.Manager, d.PlatformType = "dev-1", "Pixel", "android_device", models.PlatformTypeAndroid
	})
	master, err := wiring.BuildMaster(cfg, wiring.WithDB(db), wiring.WithLogger(quietLogger()), wiring.WithDeviceHandlers(phoneHandler()))
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(master.Server.Handler())
	t.Cleanup(server.Close)
	jar, _ := cookiejar.New(nil)
	env := &masterEnv{master: master, server: server, client: &http.Client{Jar: jar}}
	env.register(t)
	return env
}

// register signs a user up; the session cookie lands in the client's jar.
func (e *masterEnv) register(t *testing.T) {
	t.Helper()
	body := `{"name":"Ada","email":"ada@example.com","password":"long enough"}`
	res, err := e.client.Post(e.server.URL+"/auth/register", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("register: %d", res.StatusCode)
	}
}

func (e *masterEnv) get(t *testing.T, path string) (int, string) {
	t.Helper()
	res, err := e.client.Get(e.server.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	var out strings.Builder
	_, _ = bufio.NewReader(res.Body).WriteTo(&out)
	return res.StatusCode, out.String()
}

func TestMasterServesTheProfileOfTheSignedInUser(t *testing.T) {
	e := newMasterEnv(t)

	code, body := e.get(t, "/api/me")

	if code != http.StatusOK || !strings.Contains(body, "Ada") {
		t.Fatalf("%d %s", code, body)
	}
}

func TestMasterGuardsTheApiWithoutASession(t *testing.T) {
	e := newMasterEnv(t)

	res, err := http.Get(e.server.URL + "/api/me")
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestMasterServesTheProjectRoutes(t *testing.T) {
	e := newMasterEnv(t)

	code, body := e.get(t, "/api/projects")

	if code != http.StatusOK || !strings.Contains(body, `"identifier":"proj"`) {
		t.Fatalf("%d %s", code, body)
	}
	if code, _ := e.get(t, "/api/nope/devices"); code != http.StatusNotFound {
		t.Fatalf("unknown project: %d", code)
	}
}

func TestMasterServesTheMCPToolsOnItsServices(t *testing.T) {
	e := newMasterEnv(t)
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).
		Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: e.server.URL + "/api/proj/mcp", HTTPClient: e.client}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_devices"})

	if err != nil || res.IsError {
		t.Fatalf("list_devices: %v %+v", err, res)
	}
	if text := res.Content[0].(*mcp.TextContent).Text; !strings.Contains(text, `"name": "Pixel"`) {
		t.Fatalf("devices: %s", text)
	}
}

// A lock change travels through the listener of the wiring, the device
// store, the SSE bridge and the broker to a connected client.
func TestMasterStreamsDeviceChangesToSSEClients(t *testing.T) {
	e := newMasterEnv(t)
	res, err := e.client.Get(e.server.URL + "/api/sse")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	lines := bufio.NewScanner(res.Body)
	if !lines.Scan() || lines.Text() != ": connected" {
		t.Fatalf("first line %q", lines.Text())
	}

	events.DeviceLockChanged.Trigger(events.DeviceLockChangedPayload{DeviceIdentifier: "dev-1", Locked: true})

	if got := readUntil(lines, "event:device_list"); !got {
		t.Fatal("no device_list event")
	}
}

func readUntil(lines *bufio.Scanner, want string) bool {
	done := make(chan bool, 1)
	go func() {
		for lines.Scan() {
			if lines.Text() == want {
				done <- true
				return
			}
		}
		done <- false
	}()
	select {
	case ok := <-done:
		return ok
	case <-time.After(5 * time.Second):
		return false
	}
}

func TestMasterAddsTheFakeHandlerToItsDeviceManager(t *testing.T) {
	e := newMasterEnv(t)

	dev, manager := e.master.Devices.GetDevice("dev-1")

	if dev == nil || manager != "android_device" {
		t.Fatalf("device %v of %q", dev, manager)
	}
}

func TestMasterRunReturnsTheListenError(t *testing.T) {
	cfg := testConfig(t)
	cfg.Port = busyPort(t)
	master, err := wiring.BuildMaster(cfg, wiring.WithDB(testkit.DB(t)), wiring.WithLogger(quietLogger()))
	if err != nil {
		t.Fatal(err)
	}

	if err := master.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "address already in use") {
		t.Fatalf("err = %v", err)
	}
}
