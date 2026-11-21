package api

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
)

var updateSSEGolden = flag.Bool("update-sse", false, //nolint:gochecknoglobals // test flag
	"rewrite the SSE golden files in testdata/sse")

// registerAllSSE registers every event-to-SSE listener on the publisher.
func registerAllSSE(p eventRecorder) { RegisterSSEBridges(p) }

type sseCase struct {
	name    string
	trigger func()
	want    int
}

func sseCases() []sseCase {
	dev := &models.Device{DeviceIdentifier: "phone", Name: "Pixel"}
	dev.Dev = &driver{}
	protocol := &models.TestProtocol{TestRunID: 7, TestName: "Suite/Test", Device: dev}
	protocol.ID = 3
	return []sseCase{
		{"device_status", func() {
			events.DeviceStatusChanged.Trigger(events.DeviceStatusChangedPayload{DeviceID: 4, DeviceState: 2})
		}, 1},
		{"device_list", func() {
			events.DeviceListChanged.Trigger(events.DeviceListChangedPayload{DeviceID: 4, Added: true})
		}, 1},
		{"app_created", func() {
			events.AppCreated.Trigger(events.AppCreatedPayload{AppID: 5, App: map[string]any{"name": "app"}})
		}, 1},
		{"test_created", func() {
			events.TestCreated.Trigger(events.TestCreatedPayload{TestID: 6, Test: map[string]any{"name": "t"}})
		}, 1},
		{"new_test_protocol", func() {
			events.NewTestProtocol.Trigger(events.NewTestProtocolPayload{TestRunID: 7, Protocol: protocol})
		}, 1},
		{"new_test_protocol_plain", func() {
			events.NewTestProtocol.Trigger(events.NewTestProtocolPayload{TestRunID: 8, Protocol: "raw"})
		}, 1},
		{"new_test_protocol_log", func() {
			events.NewTestProtocolLog.Trigger(events.NewTestProtocolLogPayload{TestProtocolID: 9, Entry: "line"})
		}, 1},
		{"new_test_log_entry", func() {
			events.NewTestLogEntry.Trigger(events.NewTestLogEntryPayload{TestRunID: 10, Entry: "log"})
		}, 1},
		{"protocol_key", func() {
			events.ProtocolKey.Trigger(events.ProtocolKeyPayload{TestRunID: 11, TestProtocolID: 12, Key: "k"})
		}, 1},
		{"test_run_started", func() {
			events.TestRunStarted.Trigger(events.TestRunStartedPayload{ProjectID: "p", AppID: 2, TestID: 3, TestRunID: 4101})
		}, 1},
		{"test_run_finished", func() {
			events.TestRunFinished.Trigger(events.TestRunFinishedPayload{ProjectID: "p", AppID: 2, TestRunID: 4102, TestName: "n", Success: true, Succeeded: 1, TestRun: "run"})
		}, 2},
		{"node_connected", func() {
			events.NodeConnected.Trigger(events.NodeConnectedPayload{NodeIdentifier: "node-1", Hostname: "host"})
		}, 1},
		{"node_disconnected", func() {
			events.NodeDisconnected.Trigger(events.NodeDisconnectedPayload{NodeIdentifier: "node-1"})
		}, 1},
	}
}

// collectSSE waits for n published events and renders them as "channel json"
// lines sorted by channel, because handlers run in their own goroutines.
func collectSSE(t *testing.T, r eventRecorder, n int) string {
	t.Helper()
	var lines []string
	deadline := time.After(2 * time.Second)
	for len(lines) < n {
		select {
		case e := <-r:
			data, err := json.Marshal(e.Content)
			if err != nil {
				t.Fatalf("channel %s does not encode: %v", e.Channel, err)
			}
			lines = append(lines, fmt.Sprintf("%s %s", e.Channel, data))
		case <-deadline:
			t.Fatalf("got %d of %d events: %v", len(lines), n, lines)
		}
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

// The channel names and JSON bodies sent to SSE clients must not change.
func TestSSEEventsMatchGolden(t *testing.T) {
	for _, c := range sseCases() {
		t.Run(c.name, func(t *testing.T) {
			r := make(eventRecorder, 16)
			registerAllSSE(r)
			c.trigger()
			got := collectSSE(t, r, c.want)

			assertGolden(t, filepath.Join("testdata", "sse", c.name+".golden"), got)
		})
	}
}

// assertGolden compares got with the golden file, which -update-sse rewrites.
func assertGolden(t *testing.T, path, got string) {
	t.Helper()
	if *updateSSEGolden {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("SSE output differs from %s\n got: %s\nwant: %s", path, got, want)
	}
}
