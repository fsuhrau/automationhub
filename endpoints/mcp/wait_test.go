package mcp

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/hub/runs"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// failingProgress is the run service whose progress cannot be read.
type failingProgress struct{ RunService }

func (failingProgress) Progress(context.Context, uint, uint) (runsvc.Progress, error) {
	return runsvc.Progress{}, errBoom
}

// waitClient connects with a handler that reports progress every
// millisecond and collects the progress notifications.
func waitClient(t *testing.T, w *goldenWorld, h *Handler) (*mcp.ClientSession, chan *mcp.ProgressNotificationParams) {
	t.Helper()
	got := make(chan *mcp.ProgressNotificationParams, 100)
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, &mcp.ClientOptions{
		ProgressNotificationHandler: func(_ context.Context, req *mcp.ProgressNotificationClientRequest) { got <- req.Params },
	})
	cs, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint: w.serve(t, h).URL, HTTPClient: &http.Client{Transport: callerTransport{"alice"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs, got
}

func waitParams(args map[string]any) *mcp.CallToolParams {
	params := &mcp.CallToolParams{Name: "wait_for_test_run", Arguments: args}
	params.SetProgressToken("tok")
	return params
}

func TestWaitReportsProgressAndEndsWhenTheRunFinishes(t *testing.T) {
	w := newGoldenWorld(t)
	h := NewHandler(Deps{Runs: w.runs, Protocols: w.runs, Lookup: w.runs, Apps: w.apps, Tests: w.tests, Devices: w.devices,
		Pool: w.pool, Control: w.control, Tracker: w.tracker, DataDir: w.dataDir, StoreDir: w.storeDir, ProgressEvery: time.Millisecond})
	cs, got := waitClient(t, w, h)
	w.tracker.Started(1)
	done := make(chan *mcp.CallToolResult, 1)
	go func() {
		res, err := cs.CallTool(context.Background(), waitParams(a("runId", 1, "timeoutSeconds", 30)))
		if err != nil {
			t.Error(err)
		}
		done <- res
	}()

	select {
	case p := <-got:
		if p.Message != "2 of 3 protocols finished" || p.Progress != 2 || p.Total != 3 {
			t.Errorf("progress %+v", p)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("no progress notification")
	}
	w.tracker.Finished(1, runs.RunResult{Success: true})
	select {
	case res := <-done:
		if res.IsError || !strings.Contains(text(res), `"state": "finished"`) {
			t.Fatalf("result %v %s", res.IsError, text(res))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("wait did not end")
	}
}

func TestWaitGoesOnWhenTheProgressCannotBeRead(t *testing.T) {
	w := newGoldenWorld(t)
	h := NewHandler(Deps{Runs: failingProgress{w.runs}, Protocols: w.runs, Lookup: w.runs, Apps: w.apps, Tests: w.tests, Devices: w.devices,
		Pool: w.pool, Control: w.control, Tracker: w.tracker, ProgressEvery: time.Millisecond})
	w.tracker.Started(1)
	cs, got := waitClient(t, w, h)

	res, err := cs.CallTool(context.Background(), waitParams(a("runId", 1, "timeoutSeconds", 1)))

	if err != nil || res.IsError || !strings.Contains(text(res), "still running after 1s") || len(got) != 0 {
		t.Fatalf("result %v %v, %d notifications", res, err, len(got))
	}
}
