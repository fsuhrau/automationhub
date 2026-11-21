package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/hub/runs"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RunTestArgs struct {
	TestID   uint   `json:"testId" jsonschema:"test id from list_tests"`
	BundleID uint   `json:"bundleId,omitempty" jsonschema:"app bundle id from list_bundles; required except for web and editor apps"`
	StartURL string `json:"startUrl,omitempty" jsonschema:"start URL for web apps"`
	Params   string `json:"params,omitempty" jsonschema:"environment parameters as key=value;key2=value2"`
}

type RunArgs struct {
	RunID uint `json:"runId" jsonschema:"test run id"`
}

func (t *tools) registerRuns(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "run_test", Description: "Start a test run on the devices of its configuration. Returns the run id immediately; use wait_for_test_run and get_test_run.", Annotations: writeSafe()}, t.runTest)
	mcp.AddTool(s, &mcp.Tool{Name: "get_test_run", Description: "Status and summary of a test run: state, results per protocol (one per device and test), and the last run log lines.", Annotations: readOnly()}, t.getTestRun)
	mcp.AddTool(s, &mcp.Tool{Name: "wait_for_test_run", Description: "Wait until a test run finished or the timeout passed. Sends progress notifications.", Annotations: readOnly()}, t.waitForTestRun)
	mcp.AddTool(s, &mcp.Tool{Name: "cancel_test_run", Description: "Cancel a running test run.", Annotations: destructive()}, t.cancelTestRun)
}

func (t *tools) runTest(ctx context.Context, req *mcp.CallToolRequest, in RunTestArgs) (*mcp.CallToolResult, any, error) {
	ref, err := t.Runs.Start(ctx, runsvc.StartRequest{
		Project: t.project, TestID: in.TestID, BundleID: in.BundleID, StartURL: in.StartURL, Params: in.Params,
	})
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{"runId": ref.Run.ID, "sessionId": ref.Run.SessionID, "state": runs.RunRunning})
}

func (t *tools) cancelTestRun(ctx context.Context, req *mcp.CallToolRequest, in RunArgs) (*mcp.CallToolResult, any, error) {
	if err := t.Runs.Cancel(ctx, t.project.ID, in.RunID); err != nil {
		return nil, nil, err
	}
	return textResult("test run %d cancelled", in.RunID)
}

func (t *tools) getTestRun(ctx context.Context, req *mcp.CallToolRequest, in RunArgs) (*mcp.CallToolResult, any, error) {
	summary, err := t.runSummary(ctx, in.RunID)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(summary)
}

// runSummary reports a run of the project together with the state the
// tracker knows of it.
func (t *tools) runSummary(ctx context.Context, runID uint) (map[string]any, error) {
	report, err := t.Runs.Report(ctx, t.project.ID, runID)
	if err != nil {
		return nil, err
	}
	state, result := t.Tracker.Status(runID)
	summary := summarizeRun(report)
	summary["state"] = state
	if result != nil {
		summary["result"] = result
	}
	return summary, nil
}

// summarizeRun is the part of the run summary that comes from the stored
// run.
func summarizeRun(report *runsvc.Report) map[string]any {
	run := report.Run
	counts := map[string]int{}
	for i := range run.Protocols {
		counts[resultNames()[run.Protocols[i].TestResult]]++
	}
	summary := map[string]any{
		"runId":        run.ID,
		"createdAt":    run.CreatedAt,
		"parameter":    run.Parameter,
		"resultCounts": counts,
		"protocols":    toProtocolInfos(run.Protocols),
		"log":          logLines(report.Log),
	}
	if run.Test != nil {
		summary["test"] = map[string]any{"id": run.Test.ID, "name": run.Test.Name}
	}
	if run.AppBinary != nil {
		summary["bundle"] = toBundleInfo(run.AppBinary)
	}
	return summary
}

func logLines(entries []models.TestRunLogEntry) []string {
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s [%s] %s", e.CreatedAt.Format(time.RFC3339), e.Level, e.Log))
	}
	return lines
}
