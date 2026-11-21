package mcp

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var errBoom = errors.New("boom")

// stub answers every service call with an empty success, except the calls
// named in failAt: method name to the number of the call that fails (1 is
// the first).
type stub struct {
	mu     sync.Mutex
	failAt map[string]int
	calls  map[string]int
}

func (s *stub) err(method string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.calls == nil {
		s.calls = map[string]int{}
	}
	s.calls[method]++
	if s.calls[method] == s.failAt[method] {
		return errBoom
	}
	return nil
}

func (s *stub) Start(context.Context, runsvc.StartRequest) (runsvc.RunRef, error) {
	return runsvc.RunRef{Run: &models.TestRun{}}, s.err("Start")
}
func (s *stub) Cancel(context.Context, uint, uint) error { return s.err("Cancel") }
func (s *stub) Report(context.Context, uint, uint) (*runsvc.Report, error) {
	return &runsvc.Report{Run: &models.TestRun{}}, s.err("Report")
}
func (s *stub) Progress(context.Context, uint, uint) (runsvc.Progress, error) {
	return runsvc.Progress{}, s.err("Progress")
}
func (s *stub) ProtocolDetail(context.Context, uint, uint) (*runsvc.ProtocolDetail, error) {
	return &runsvc.ProtocolDetail{Protocol: &models.TestProtocol{}}, s.err("ProtocolDetail")
}
func (s *stub) ProtocolData(context.Context, uint, uint) ([]models.ProtocolEntry, error) {
	return nil, s.err("ProtocolData")
}
func (s *stub) Logs(context.Context, runsvc.LogQuery) (*runsvc.LogPage, error) {
	return &runsvc.LogPage{}, s.err("Logs")
}
func (s *stub) Devices(context.Context, *models.Project) ([]models.Device, error) {
	return []models.Device{{}}, s.err("Devices")
}
func (s *stub) Device(context.Context, *models.Project, uint) (*models.Device, error) {
	return &models.Device{DeviceIdentifier: "dev-1", Name: "Pixel"}, s.err("Device")
}
func (s *stub) App(context.Context, appsvc.Scope) (*models.App, error) {
	return &models.App{}, s.err("App")
}
func (s *stub) Apps(context.Context, uint) ([]models.App, error) { return nil, s.err("Apps") }
func (s *stub) Bundles(context.Context, appsvc.Scope) ([]appsvc.BundleEntry, error) {
	return nil, s.err("Bundles")
}
func (s *stub) StoreBundle(context.Context, appsvc.Scope, string) (*models.AppBinary, error) {
	return &models.AppBinary{}, s.err("StoreBundle")
}
func (s *stub) List(context.Context, testsvc.Scope) ([]testsvc.ListEntry, error) {
	return nil, s.err("List")
}
func (s *stub) ControlApp(context.Context, *models.Project, uint) (*app.Parameter, error) {
	return &app.Parameter{}, s.err("ControlApp")
}
func (s *stub) Unlock(context.Context, *models.Project, string) (*models.Device, error) {
	return &models.Device{}, s.err("Unlock")
}

// stubbed is the world's handler with every service replaced by a stub.
func (w *goldenWorld) stubbed(failAt map[string]int) *Handler {
	s := &stub{failAt: failAt}
	return NewHandler(Deps{
		Runs: s, Protocols: s, Lookup: s, Apps: s, Tests: s, Devices: s,
		Pool: w.pool, Control: w.control, Tracker: w.tracker, DataDir: w.dataDir, StoreDir: w.storeDir,
	})
}

func TestEveryServiceFailureIsAToolError(t *testing.T) {
	lock := step{"lock_device", a("deviceId", 1)}
	tests := []struct {
		method string
		at     int
		steps  []step
	}{
		{"Devices", 1, []step{{"list_devices", a()}}},
		{"Device", 1, []step{{"get_device", a("deviceId", 1)}}},
		{"Unlock", 1, []step{{"unlock_device", a("deviceId", 1)}}},
		{"Apps", 1, []step{{"list_apps", a()}}},
		{"App", 1, []step{{"list_bundles", a("appId", 1)}}},
		{"Bundles", 1, []step{{"list_bundles", a("appId", 1)}}},
		{"App", 1, []step{{"list_tests", a("appId", 1)}}},
		{"List", 1, []step{{"list_tests", a("appId", 1)}}},
		{"App", 1, []step{{"upload_bundle", a("appId", 1, "url", "{files}/builds/ok.apk")}}},
		{"StoreBundle", 1, []step{{"upload_bundle", a("appId", 1, "url", "{files}/builds/ok.apk")}}},
		{"App", 1, []step{{"begin_upload", a("appId", 1, "fileName", "a.apk")}}},
		{"Start", 1, []step{{"run_test", a("testId", 1)}}},
		{"Cancel", 1, []step{{"cancel_test_run", a("runId", 1)}}},
		{"Report", 1, []step{{"get_test_run", a("runId", 1)}}},
		{"Report", 1, []step{{"wait_for_test_run", a("runId", 1)}}},
		{"Report", 2, []step{{"wait_for_test_run", a("runId", 1)}}},
		{"ProtocolDetail", 1, []step{{"get_protocol", a("protocolId", 1)}}},
		{"ProtocolData", 1, []step{{"get_protocol_screenshot", a("protocolId", 1)}}},
		{"Logs", 1, []step{{"get_protocol_logs", a("protocolId", 1)}}},
		{"ControlApp", 1, []step{lock, {"install_app", a("deviceId", 1, "bundleId", 1)}}},
		{"ControlApp", 1, []step{lock, {"start_app", a("deviceId", 1, "bundleId", 1)}}},
		{"ControlApp", 1, []step{lock, {"stop_app", a("deviceId", 1, "bundleId", 1)}}},
	}
	for _, tc := range tests {
		t.Run(tc.method+"/"+tc.steps[len(tc.steps)-1].tool, func(t *testing.T) {
			w := newBareWorld(t)
			cs := connect(t, w.serve(t, w.stubbed(map[string]int{tc.method: tc.at})), "alice")

			last := runSteps(t, w, cs, tc.steps)

			if !last.IsError || text(last) != "boom" {
				t.Errorf("last result error %v %q, want the %s error", last.IsError, text(last), tc.method)
			}
		})
	}
}

// runSteps calls the tools in order and returns the result of the last.
func runSteps(t *testing.T, w *goldenWorld, cs *mcp.ClientSession, steps []step) *mcp.CallToolResult {
	t.Helper()
	var res *mcp.CallToolResult
	for _, s := range steps {
		args := map[string]any{}
		for k, v := range s.args {
			args[k] = v
			if str, ok := v.(string); ok {
				args[k] = replaceFiles(str, w)
			}
		}
		res = call(t, cs, s.tool, args)
	}
	return res
}

func TestEveryControlToolFailsWhenTheDeviceLookupFails(t *testing.T) {
	tools := map[string]map[string]any{
		"lock_device": a("deviceId", 1), "release_device": a("deviceId", 1), "screenshot": a("deviceId", 1),
		"get_hierarchy": a("deviceId", 1), "device_logs": a("deviceId", 1), "tap": a("deviceId", 1, "x", 1, "y", 1),
		"swipe": a("deviceId", 1, "fromX", 1, "fromY", 1, "toX", 2, "toY", 2), "type_text": a("deviceId", 1, "text", "x"),
		"press_key":   a("deviceId", 1, "key", "home"),
		"install_app": a("deviceId", 1, "bundleId", 1), "start_app": a("deviceId", 1, "bundleId", 1),
		"stop_app": a("deviceId", 1, "bundleId", 1),
	}
	for name, args := range tools {
		t.Run(name, func(t *testing.T) {
			w := newBareWorld(t)
			cs := connect(t, w.serve(t, w.stubbed(map[string]int{"Device": 1})), "alice")

			res := call(t, cs, name, args)

			if !res.IsError || text(res) != "boom" {
				t.Fatalf("%s: %v %q, want the lookup error", name, res.IsError, text(res))
			}
		})
	}
}
