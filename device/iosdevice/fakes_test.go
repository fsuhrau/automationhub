package iosdevice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/modules/webdriver"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
)

const wdaURLLine = "ServerURLHere->http://10.0.0.2:8100<-ServerURLHere\n"

// wdaServer is an httptest WebDriverAgent that records "METHOD path body".
type wdaServer struct {
	*httptest.Server
	mu     sync.Mutex
	routes map[string]string
	seen   []string
}

func newWDAServer(t *testing.T) *wdaServer {
	t.Helper()
	f := &wdaServer{routes: map[string]string{}}
	for route, answer := range wdaAnswers() {
		f.routes[route] = answer
	}
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.Close)
	return f
}

func wdaAnswers() map[string]string {
	return map[string]string{
		"GET /status":                                `{"value":{"ready":true}}`,
		"POST /session":                              `{"sessionId":"s1","value":{"sessionId":"s1","capabilities":{}}}`,
		"DELETE /session":                            `{"value":null}`,
		"GET /wda/screen":                            `{"value":{"screenSize":{"width":100,"height":200},"scale":2}}`,
		"POST /session/s1/wda/apps/launch":           `{"value":null}`,
		"POST /session/s1/wda/apps/terminate":        `{"value":true}`,
		"GET /session/s1/wda/activeAppInfo":          `{"value":{"bundleId":"com.app"}}`,
		"POST /session/s1/wda/tap":                   `{"value":null}`,
		"POST /session/s1/wda/dragfromtoforduration": `{"value":null}`,
		"POST /session/s1/wda/keys":                  `{"value":null}`,
		"POST /session/s1/wda/pressButton":           `{"value":null}`,
		"POST /session/s1/element":                   `{"value":{"ELEMENT":"e1"}}`,
		"POST /session/s1/element/e1/click":          `{"value":null}`,
		"GET /session/s1/alert/text":                 `{"value":"Allow?"}`,
	}
}

func (f *wdaServer) serve(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	f.seen = append(f.seen, fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, b))
	answer, ok := f.routes[r.Method+" "+r.URL.Path]
	f.mu.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		answer = `{"value":{"error":"unknown command","message":"no such route"}}`
	}
	_, _ = io.WriteString(w, answer)
}

// remove makes a route answer 404 like an unknown command.
func (f *wdaServer) remove(route string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.routes, route)
}

// requests are the recorded requests, "METHOD path body".
func (f *wdaServer) requests() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.seen...)
}

func (f *wdaServer) saw(t *testing.T, want string) {
	t.Helper()
	for _, r := range f.requests() {
		if strings.HasPrefix(r, want) {
			return
		}
	}
	t.Errorf("WebDriverAgent never saw %q; requests: %q", want, f.requests())
}

func (f *wdaServer) hostPort() string { return strings.TrimPrefix(f.URL, "http://") }

// fakeForward is a Forwarder to a fixed address.
type fakeForward struct {
	addr   string
	closed *int
}

func (f fakeForward) Addr() string { return f.addr }
func (f fakeForward) Close()       { *f.closed++ }

// fakeUSB is the go-ios side of a host with attached devices.
type fakeUSB struct {
	mu        sync.Mutex
	serials   []string
	locked    map[string]bool // serials lockdown refuses
	listErr   error
	lists     int
	probeErr  error
	forwardTo string // address of the fake WebDriverAgent
	forwarded []uint16
	closed    int
	xcuiErr   error
	xcuiFn    func() error // replaces xcuiErr when set
	xcuiRuns  []string
	xcuiClose int
}

func (u *fakeUSB) ops() USBOps {
	return USBOps{
		ListDevices: u.listDevices,
		GetValues:   u.getValues,
		GetDevice:   u.getDevice,
		Probe:       func(ios.DeviceEntry, uint16) error { return u.probeErr },
		Forward:     u.forward,
		RunXCUI:     u.runXCUI,
		CloseXCUI:   func() error { u.xcuiClose++; return nil },
	}
}

func entryOf(serial string) ios.DeviceEntry {
	return ios.DeviceEntry{DeviceID: 7, Properties: ios.DeviceProperties{SerialNumber: serial}}
}

func (u *fakeUSB) listDevices() (ios.DeviceList, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.lists++
	var list ios.DeviceList
	for _, s := range u.serials {
		list.DeviceList = append(list.DeviceList, entryOf(s))
	}
	return list, u.listErr
}

func (u *fakeUSB) getValues(e ios.DeviceEntry) (ios.GetAllValuesResponse, error) {
	serial := e.Properties.SerialNumber
	if u.locked[serial] {
		return ios.GetAllValuesResponse{}, errors.New("lockdown: pairing required")
	}
	var r ios.GetAllValuesResponse
	r.Value.DeviceName = "iPhone of " + serial
	r.Value.ProductName = "iPhone OS"
	r.Value.ProductVersion = "18.1"
	r.Value.ProductType = "iPhone15,2"
	return r, nil
}

func (u *fakeUSB) getDevice(udid string) (ios.DeviceEntry, error) {
	for _, s := range u.serials {
		if s == udid {
			return entryOf(s), nil
		}
	}
	return ios.DeviceEntry{}, fmt.Errorf("device %s not found", udid)
}

func (u *fakeUSB) forward(_ ios.DeviceEntry, port uint16) (Forwarder, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.forwarded = append(u.forwarded, port)
	if port == 9100 {
		return fakeForward{addr: "127.0.0.1:9100", closed: &u.closed}, nil
	}
	return fakeForward{addr: u.forwardTo, closed: &u.closed}, nil
}

func (u *fakeUSB) runXCUI(bundle, _, config string, _ ios.DeviceEntry, _, _ []string) error {
	u.mu.Lock()
	u.xcuiRuns = append(u.xcuiRuns, bundle+" "+config)
	u.mu.Unlock()
	if u.xcuiFn != nil {
		return u.xcuiFn()
	}
	return u.xcuiErr
}

// fakeProc is the xcodebuild process of fakeStarter.
type fakeProc struct {
	wrote    chan struct{}
	block    chan struct{}
	once     sync.Once
	stubborn bool // ignores Interrupt, only Kill ends it
	signals  []string
	mu       sync.Mutex
}

func (p *fakeProc) end() { p.once.Do(func() { close(p.block) }) }

func (p *fakeProc) Wait() error { <-p.wrote; <-p.block; return nil }

func (p *fakeProc) Interrupt() error {
	p.record("interrupt")
	if !p.stubborn {
		p.end()
	}
	return nil
}

func (p *fakeProc) Kill() error { p.record("kill"); p.end(); return nil }

func (p *fakeProc) record(s string) { p.mu.Lock(); p.signals = append(p.signals, s); p.mu.Unlock() }

func (p *fakeProc) got() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.signals...)
}

// fakeStarter fakes xcodebuild. Commands are recorded and checked by a
// testkit.FakeRunner: an unscripted command line is a start error.
type fakeStarter struct {
	runner   testkit.FakeRunner
	output   string // xcodebuild output
	exits    bool   // xcodebuild ends right after its output
	stubborn bool
	proc     *fakeProc
}

func (s *fakeStarter) StartStreaming(out io.Writer, name string, args ...string) (RunnerProcess, error) {
	if _, err := s.runner.Start(context.Background(), name, args...); err != nil {
		return nil, err
	}
	p := &fakeProc{wrote: make(chan struct{}), block: make(chan struct{}), stubborn: s.stubborn}
	if s.exits {
		p.end()
	}
	go func() {
		_, _ = io.WriteString(out, s.output)
		close(p.wrote)
	}()
	s.proc = p
	return p, nil
}

// fakeStore is the storage.Device of the handler.
type fakeStore struct {
	mu      sync.Mutex
	stored  models.Devices
	created []models.Device
	updates []string // "id state"
}

func (s *fakeStore) GetDevices(string) (models.Devices, error) { return s.stored, nil }
func (s *fakeStore) GetDevice(string, string) (*models.Device, error) {
	return nil, errors.New("not used")
}
func (s *fakeStore) NewDevice(_ string, d models.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.created = append(s.created, d)
	return nil
}
func (s *fakeStore) Update(_ string, d device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updates = append(s.updates, d.DeviceID()+" "+device.StateToString(d.DeviceState()))
	return nil
}

func (s *fakeStore) lastUpdates(n int) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.updates) < n {
		return append([]string(nil), s.updates...)
	}
	return append([]string(nil), s.updates[len(s.updates)-n:]...)
}

func t0() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }

// rig is a handler on fakes.
type rig struct {
	h     *Handler
	usb   *fakeUSB
	proc  *fakeStarter
	clock *testkit.FakeClock
	store *fakeStore
	wda   *wdaServer
	log   *logrus.Logger
	hook  *webdriver.WDAConnectHook
}

const xctestrun = "/wda/WDA.xctestrun"

func newRig(t *testing.T, serials ...string) *rig {
	t.Helper()
	return newRigWith(t, &config.WebDriver{XCTestRun: xctestrun}, serials...)
}

// newRigWith is newRig with another WebDriverAgent config.
func newRigWith(t *testing.T, cfg *config.WebDriver, serials ...string) *rig {
	t.Helper()
	r := &rig{usb: &fakeUSB{serials: serials}, proc: &fakeStarter{}, clock: testkit.NewFakeClock(t0()), store: &fakeStore{}, wda: newWDAServer(t)}
	r.log = logrus.New()
	r.log.SetOutput(io.Discard)
	r.hook = &webdriver.WDAConnectHook{}
	r.log.AddHook(r.hook)
	r.usb.forwardTo = r.wda.hostPort()
	r.usb.probeErr = errors.New("nothing listens")
	r.proc.output = wdaURLLine
	for _, s := range serials {
		r.proc.runner.Script("xcodebuild test-without-building -xctestrun "+xctestrun+" -destination id="+s, nil, nil)
	}
	deps := Deps{Clock: r.clock, Processes: r.proc, USB: r.usb.ops(), Hook: r.hook}
	r.h = NewHandlerWithDeps(config.Manager{WebDriver: cfg}, r.store, deps)
	r.h.masterURL = strings.TrimPrefix(newMaster(t), "http://")
	return r
}

func (r *rig) device(t *testing.T, serial string) *Device {
	t.Helper()
	d, ok := r.h.Get(serial)
	if !ok {
		t.Fatalf("device %s unknown", serial)
	}
	return d
}

// newMaster is a master that accepts the logs and actions of the devices.
func newMaster(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(srv.Close)
	return srv.URL
}
