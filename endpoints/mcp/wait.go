package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultWait = 5 * time.Minute
	maxWait     = 30 * time.Minute
)

type WaitArgs struct {
	RunID          uint `json:"runId" jsonschema:"test run id"`
	TimeoutSeconds int  `json:"timeoutSeconds,omitempty" jsonschema:"how long to wait (default 300, max 1800)"`
}

// waitTimeout is how long wait_for_test_run waits: the default for none, at
// most maxWait.
func waitTimeout(seconds int) time.Duration {
	timeout := time.Duration(seconds) * time.Second
	if timeout <= 0 {
		return defaultWait
	}
	return min(timeout, maxWait)
}

func (t *tools) waitForTestRun(ctx context.Context, req *mcp.CallToolRequest, in WaitArgs) (*mcp.CallToolResult, any, error) {
	if _, err := t.runSummary(ctx, in.RunID); err != nil {
		return nil, nil, err
	}
	timeout := waitTimeout(in.TimeoutSeconds)
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if token := req.Params.GetProgressToken(); token != nil && req.Session != nil {
		go t.reportProgress(waitCtx, progressTarget{req.Session, token, in.RunID})
	}

	state, _, err := t.Tracker.Wait(waitCtx, in.RunID)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return nil, nil, err
	}
	summary, err := t.runSummary(ctx, in.RunID)
	if err != nil {
		return nil, nil, err
	}
	if state == runs.RunRunning {
		summary["note"] = fmt.Sprintf("still running after %v; call wait_for_test_run again", timeout)
	}
	return jsonResult(summary)
}

// progressTarget is the client session that asked for the progress of a run.
type progressTarget struct {
	session *mcp.ServerSession
	token   any
	runID   uint
}

// reportProgress tells the client every ProgressEvery how many protocols of
// the run ended, until ctx is done.
func (t *tools) reportProgress(ctx context.Context, to progressTarget) {
	ticker := time.NewTicker(t.ProgressEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			progress, err := t.Runs.Progress(ctx, t.project.ID, to.runID)
			if err != nil {
				continue
			}
			_ = to.session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
				ProgressToken: to.token,
				Progress:      float64(progress.Ended),
				Total:         float64(progress.Total),
				Message:       fmt.Sprintf("%d of %d protocols finished", progress.Ended, progress.Total),
			})
		}
	}
}
