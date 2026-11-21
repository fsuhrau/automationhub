package mcp

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMain registers the -update flag: it rewrites the golden files of the
// tool results.
func TestMain(m *testing.M) {
	flag.Bool("update", false, "rewrite the golden files of the tool results")
	flag.Parse()
	os.Exit(m.Run())
}

func updateGolden() bool { return flag.Lookup("update").Value.String() == "true" }

// step is one tool call; {files} in a string argument is the address of the
// file server that serves builds.
type step struct {
	tool string
	args map[string]any
}

type goldenCase struct {
	name  string
	steps []step
}

// goldenOpts says how a group of cases runs: with fresh each case gets its
// own world, otherwise they share one; with app the device has a connected
// app that answers with goldenScene and a screenshot.
type goldenOpts struct{ fresh, app bool }

const goldenScene = `<Root ID="1" Name="root"><Button ID="8" Name="Play"/><Button ID="9" Name="Quit"/></Root>`

func call1(tool string, args map[string]any) []step { return []step{{tool, args}} }

func a(kv ...any) map[string]any {
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}

func orchestrationCases() []goldenCase {
	return []goldenCase{
		{"list_devices", call1("list_devices", a())},
		{"list_devices_android", call1("list_devices", a("platform", "Android"))},
		{"list_devices_unknown_platform", call1("list_devices", a("platform", "toaster"))},
		{"get_device", call1("get_device", a("deviceId", 1))},
		{"get_device_offline_node", call1("get_device", a("deviceId", 2))},
		{"get_device_unknown", call1("get_device", a("deviceId", 42))},
		{"list_apps", call1("list_apps", a())},
		{"list_bundles", call1("list_bundles", a("appId", 1))},
		{"list_bundles_limit", call1("list_bundles", a("appId", 1, "limit", 1))},
		{"list_bundles_unknown_app", call1("list_bundles", a("appId", 99))},
		{"list_bundles_foreign_app", call1("list_bundles", a("appId", 3))},
		{"list_tests", call1("list_tests", a("appId", 1))},
		{"list_tests_unknown_app", call1("list_tests", a("appId", 99))},
		{"list_tests_foreign_app", call1("list_tests", a("appId", 3))},
		{"get_test_run", call1("get_test_run", a("runId", 1))},
		{"get_test_run_unknown", call1("get_test_run", a("runId", 99))},
		{"get_test_run_foreign", call1("get_test_run", a("runId", 2))},
		{"get_protocol", call1("get_protocol", a("protocolId", 1))},
		{"get_protocol_child", call1("get_protocol", a("protocolId", 2))},
		{"get_protocol_unknown", call1("get_protocol", a("protocolId", 99))},
		{"get_protocol_foreign", call1("get_protocol", a("protocolId", 4))},
		{"get_protocol_logs", call1("get_protocol_logs", a("protocolId", 1))},
		{"get_protocol_logs_level", call1("get_protocol_logs", a("protocolId", 1, "level", "ERROR"))},
		{"get_protocol_logs_contains", call1("get_protocol_logs", a("protocolId", 1, "contains", "boom"))},
		{"get_protocol_logs_page", call1("get_protocol_logs", a("protocolId", 1, "offset", 1, "limit", 2))},
		{"get_protocol_logs_percent_sign", call1("get_protocol_logs", a("protocolId", 1, "contains", "50%"))},
		{"get_protocol_logs_unknown", call1("get_protocol_logs", a("protocolId", 99))},
		{"get_protocol_logs_foreign", call1("get_protocol_logs", a("protocolId", 4))},
		{"wait_for_test_run_unknown", call1("wait_for_test_run", a("runId", 99))},
		{"wait_for_test_run_timeout", call1("wait_for_test_run", a("runId", 1, "timeoutSeconds", 1))},
		{"cancel_test_run_without_runner", call1("cancel_test_run", a("runId", 1))},
		{"cancel_test_run_unknown", call1("cancel_test_run", a("runId", 99))},
		{"cancel_test_run_foreign", call1("cancel_test_run", a("runId", 2))},
		{"unlock_device_unknown", call1("unlock_device", a("deviceId", 42))},
		{"unlock_device_not_connected", call1("unlock_device", a("deviceId", 2))},
	}
}

func runCases() []goldenCase {
	return []goldenCase{
		{"run_test", call1("run_test", a("testId", 1, "bundleId", 1, "params", "a=b"))},
		{"run_test_web_start_url", call1("run_test", a("testId", 4, "startUrl", "http://game"))},
		{"run_test_unknown_test", call1("run_test", a("testId", 99, "bundleId", 1))},
		{"run_test_foreign_test", call1("run_test", a("testId", 3, "bundleId", 1))},
		{"run_test_unknown_bundle", call1("run_test", a("testId", 1, "bundleId", 99))},
		{"run_test_archived_bundle", call1("run_test", a("testId", 1, "bundleId", 2))},
		{"run_test_without_unity_settings", call1("run_test", a("testId", 2, "bundleId", 1))},
		{"run_test_then_cancel", []step{
			{"run_test", a("testId", 1, "bundleId", 1)},
			{"cancel_test_run", a("runId", 3)},
			{"cancel_test_run", a("runId", 3)},
		}},
		{"run_test_then_wait_still_running", []step{
			{"run_test", a("testId", 1, "bundleId", 1)},
			{"wait_for_test_run", a("runId", 3, "timeoutSeconds", 1)},
		}},
		{"run_test_then_get", []step{
			{"run_test", a("testId", 1, "bundleId", 1)},
			{"get_test_run", a("runId", 3)},
			{"list_tests", a("appId", 1)},
		}},
	}
}

func uploadCases() []goldenCase {
	return []goldenCase{
		{"upload_bundle", call1("upload_bundle", a("appId", 1, "url", "{files}/builds/ok.apk"))},
		{"upload_bundle_twice", []step{
			{"upload_bundle", a("appId", 1, "url", "{files}/builds/ok.apk")},
			{"upload_bundle", a("appId", 1, "url", "{files}/builds/dup.apk")},
			{"list_bundles", a("appId", 1)},
		}},
		{"upload_bundle_bad_scheme", call1("upload_bundle", a("appId", 1, "url", "ftp://x/y.apk"))},
		{"upload_bundle_no_file_name", call1("upload_bundle", a("appId", 1, "url", "{files}/"))},
		{"upload_bundle_download_fails", call1("upload_bundle", a("appId", 1, "url", "{files}/builds/missing.apk"))},
		{"upload_bundle_unknown_app", call1("upload_bundle", a("appId", 99, "url", "{files}/builds/ok.apk"))},
		{"upload_bundle_foreign_app", call1("upload_bundle", a("appId", 3, "url", "{files}/builds/ok.apk"))},
		{"begin_upload", call1("begin_upload", a("appId", 1, "fileName", "Client-1.58.0.apk"))},
		{"begin_upload_bad_name", call1("begin_upload", a("appId", 1, "fileName", ".."))},
		{"begin_upload_unknown_app", call1("begin_upload", a("appId", 99, "fileName", "x.apk"))},
		{"begin_upload_foreign_app", call1("begin_upload", a("appId", 3, "fileName", "x.apk"))},
	}
}

func screenshotCases() []goldenCase {
	return []goldenCase{
		{"get_protocol_screenshot_last", call1("get_protocol_screenshot", a("protocolId", 1, "file", "last.png"))},
		{"get_protocol_screenshot_scaled", call1("get_protocol_screenshot", a("protocolId", 1, "file", "last.png", "maxWidth", 300))},
		{"get_protocol_screenshot_named", call1("get_protocol_screenshot", a("protocolId", 1, "file", "first.png"))},
		{"get_protocol_screenshot_default", call1("get_protocol_screenshot", a("protocolId", 1))},
		{"get_protocol_screenshot_not_a_screenshot", call1("get_protocol_screenshot", a("protocolId", 1, "file", "trace.json"))},
		{"get_protocol_screenshot_unreferenced", call1("get_protocol_screenshot", a("protocolId", 1, "file", "foreign.png"))},
		{"get_protocol_screenshot_gone", call1("get_protocol_screenshot", a("protocolId", 1, "file", "gone.png"))},
		{"get_protocol_screenshot_broken", call1("get_protocol_screenshot", a("protocolId", 1, "file", "bad.png"))},
		{"get_protocol_screenshot_none", call1("get_protocol_screenshot", a("protocolId", 3))},
		{"get_protocol_screenshot_foreign", call1("get_protocol_screenshot", a("protocolId", 4))},
		{"get_protocol_screenshot_unknown", call1("get_protocol_screenshot", a("protocolId", 99))},
	}
}

func controlCases() []goldenCase {
	return []goldenCase{
		{"lock_device", call1("lock_device", a("deviceId", 1))},
		{"lock_device_twice", []step{{"lock_device", a("deviceId", 1)}, {"lock_device", a("deviceId", 1)}}},
		{"lock_device_not_connected", call1("lock_device", a("deviceId", 2))},
		{"lock_device_unknown", call1("lock_device", a("deviceId", 42))},
		{"release_device_without_lock", call1("release_device", a("deviceId", 1))},
		{"release_device", []step{{"lock_device", a("deviceId", 1)}, {"release_device", a("deviceId", 1)}, {"release_device", a("deviceId", 1)}}},
		{"unlock_device", []step{{"lock_device", a("deviceId", 1)}, {"unlock_device", a("deviceId", 1)}, {"release_device", a("deviceId", 1)}}},
		{"screenshot", call1("screenshot", a("deviceId", 1))},
		{"screenshot_scaled", call1("screenshot", a("deviceId", 1, "maxWidth", 500))},
		{"screenshot_app_not_connected", call1("screenshot", a("deviceId", 1, "source", "app"))},
		{"screenshot_unknown", call1("screenshot", a("deviceId", 42))},
		{"screenshot_not_connected", call1("screenshot", a("deviceId", 2))},
		{"get_hierarchy_app_not_connected", call1("get_hierarchy", a("deviceId", 1))},
		{"get_hierarchy_unknown", call1("get_hierarchy", a("deviceId", 42))},
		{"tap_without_lock", call1("tap", a("deviceId", 1, "x", 5, "y", 6))},
		{"tap", []step{{"lock_device", a("deviceId", 1)}, {"tap", a("deviceId", 1, "x", 5, "y", 6)}}},
		{"tap_nothing_to_tap", []step{{"lock_device", a("deviceId", 1)}, {"tap", a("deviceId", 1)}}},
		{"tap_element_app_not_connected", []step{{"lock_device", a("deviceId", 1)}, {"tap", a("deviceId", 1, "elementId", "e1")}, {"tap", a("deviceId", 1, "xpath", "//Button")}}},
		{"swipe", []step{{"lock_device", a("deviceId", 1)}, {"swipe", a("deviceId", 1, "fromX", 1, "fromY", 2, "toX", 3, "toY", 4)}, {"swipe", a("deviceId", 1, "fromX", 1, "fromY", 2, "toX", 3, "toY", 4, "durationMs", 50)}}},
		{"swipe_without_lock", call1("swipe", a("deviceId", 1, "fromX", 1, "fromY", 2, "toX", 3, "toY", 4))},
		{"type_text", []step{{"lock_device", a("deviceId", 1)}, {"type_text", a("deviceId", 1, "text", "héllo")}}},
		{"type_text_without_lock", call1("type_text", a("deviceId", 1, "text", "x"))},
		{"press_key", []step{{"lock_device", a("deviceId", 1)}, {"press_key", a("deviceId", 1, "key", "home")}}},
		{"press_key_without_lock", call1("press_key", a("deviceId", 1, "key", "home"))},
		{"install_app", []step{{"lock_device", a("deviceId", 1)}, {"install_app", a("deviceId", 1, "bundleId", 1)}}},
		{"install_app_archived", []step{{"lock_device", a("deviceId", 1)}, {"install_app", a("deviceId", 1, "bundleId", 2)}}},
		{"install_app_unknown_bundle", []step{{"lock_device", a("deviceId", 1)}, {"install_app", a("deviceId", 1, "bundleId", 99)}}},
		{"install_app_foreign_bundle", []step{{"lock_device", a("deviceId", 1)}, {"install_app", a("deviceId", 1, "bundleId", 3)}}},
		{"install_app_without_lock", call1("install_app", a("deviceId", 1, "bundleId", 1))},
		{"start_app", []step{{"lock_device", a("deviceId", 1)}, {"start_app", a("deviceId", 1, "bundleId", 1)}}},
		{"start_app_archived", []step{{"lock_device", a("deviceId", 1)}, {"start_app", a("deviceId", 1, "bundleId", 2)}}},
		{"stop_app", []step{{"lock_device", a("deviceId", 1)}, {"stop_app", a("deviceId", 1, "bundleId", 1)}, {"stop_app", a("deviceId", 1, "bundleId", 1, "clearData", true)}}},
		{"stop_app_unknown_bundle", []step{{"lock_device", a("deviceId", 1)}, {"stop_app", a("deviceId", 1, "bundleId", 99)}}},
		{"device_logs", call1("device_logs", a("deviceId", 1))},
		{"device_logs_lines", call1("device_logs", a("deviceId", 1, "lines", 1))},
		{"device_logs_unknown", call1("device_logs", a("deviceId", 42))},
	}
}

// replaceFiles puts the address of the file server into the string.
func replaceFiles(s string, w *goldenWorld) string {
	return strings.ReplaceAll(s, "{files}", w.files.URL)
}

func appCases() []goldenCase {
	return []goldenCase{
		{"get_hierarchy", call1("get_hierarchy", a("deviceId", 1))},
		{"get_hierarchy_xpath", call1("get_hierarchy", a("deviceId", 1, "xpath", "//Button[@Name='Play']"))},
		{"get_hierarchy_invalid_xpath", call1("get_hierarchy", a("deviceId", 1, "xpath", "//["))},
		{"get_hierarchy_truncated", call1("get_hierarchy", a("deviceId", 1, "maxChars", 20))},
		{"screenshot_of_the_app", call1("screenshot", a("deviceId", 1, "source", "app"))},
		{"tap_element_of_the_app", []step{{"lock_device", a("deviceId", 1)}, {"tap", a("deviceId", 1, "elementId", "8")}, {"tap", a("deviceId", 1, "xpath", "//Button[@Name='Play']")}}},
	}
}

var (
	reTime = regexp.MustCompile(`\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(\.\d+)?(Z|[+-]\d\d:\d\d)`)
	reHex  = regexp.MustCompile(`[0-9a-f]{64}`)
	reHost = regexp.MustCompile(`127\.0\.0\.1:\d+`)
	reSess = regexp.MustCompile(`mcp:alice:[A-Z0-9]+`)
)

// normalize replaces what differs between runs: times, random ids, ports.
func normalize(s string) string {
	s = reTime.ReplaceAllString(s, "<time>")
	s = reHex.ReplaceAllString(s, "<id>")
	s = reSess.ReplaceAllString(s, "mcp:alice:<session>")
	return reHost.ReplaceAllString(s, "<host>")
}

// render writes a tool result as text: the error flag, then each content.
func render(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	var b strings.Builder
	fmt.Fprintf(&b, "isError: %v\n", res.IsError)
	for _, c := range res.Content {
		switch c := c.(type) {
		case *mcp.TextContent:
			b.WriteString("--- text\n" + c.Text + "\n")
		case *mcp.ImageContent:
			cfg, _, err := image.DecodeConfig(bytes.NewReader(c.Data))
			if err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(&b, "--- image %s %dx%d\n", c.MIMEType, cfg.Width, cfg.Height)
		default:
			t.Fatalf("unexpected content %T", c)
		}
	}
	return b.String()
}

func runCase(t *testing.T, w *goldenWorld, tc goldenCase, app bool) string {
	if app {
		w.dev.AppConnected, w.pool.scene, w.pool.screenshot = true, goldenScene, testPNG(t, 640, 480)
	}
	h := w.handler()
	cs := connect(t, w.serve(t, h), "alice")

	var out strings.Builder
	for _, s := range tc.steps {
		args := map[string]any{}
		for k, v := range s.args {
			if str, ok := v.(string); ok {
				v = replaceFiles(str, w)
			}
			args[k] = v
		}
		res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: s.tool, Arguments: args})
		if err != nil {
			t.Fatalf("%s: %v", s.tool, err)
		}
		fmt.Fprintf(&out, "=== %s %v\n%s\n", s.tool, s.args, render(t, res))
	}
	return normalize(out.String())
}

// compareGolden checks the result against the golden file of the case, or
// rewrites the file with -update.
func compareGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name+".golden")
	if updateGolden() {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if got != string(want) {
		t.Errorf("tool result changed\n--- want\n%s\n--- got\n%s", want, got)
	}
}

// checkGolden runs the cases and compares each result with its golden file.
func checkGolden(t *testing.T, cases []goldenCase, opts goldenOpts) {
	worldOf := newGoldenWorld
	if !opts.fresh {
		shared := newGoldenWorld(t)
		worldOf = func(*testing.T) *goldenWorld { return shared }
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compareGolden(t, tc.name, runCase(t, worldOf(t), tc, opts.app))
		})
	}
}

func TestGoldenReadOnlyTools(t *testing.T) {
	checkGolden(t, append(orchestrationCases(), screenshotCases()...), goldenOpts{})
}

func TestGoldenRunTools(t *testing.T)    { checkGolden(t, runCases(), goldenOpts{fresh: true}) }
func TestGoldenUploadTools(t *testing.T) { checkGolden(t, uploadCases(), goldenOpts{fresh: true}) }
func TestGoldenControlTools(t *testing.T) {
	checkGolden(t, controlCases(), goldenOpts{fresh: true})
}

func TestGoldenToolsOfAConnectedApp(t *testing.T) {
	checkGolden(t, appCases(), goldenOpts{fresh: true, app: true})
}
