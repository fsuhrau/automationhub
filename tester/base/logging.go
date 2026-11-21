package base

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
)

// LogInfo writes a line to the log of the run.
func (r *Runner) LogInfo(format string, params ...interface{}) {
	r.addLog("log", fmt.Sprintf(format, params...))
}

// LogError writes an error line to the log of the run.
func (r *Runner) LogError(format string, params ...interface{}) {
	r.addLog("error", fmt.Sprintf(format, params...))
}

func (r *Runner) addLog(level, text string) {
	entry := &models.TestRunLogEntry{TestRunID: r.TestRun.ID, Level: level, Log: text}
	// the log of a cancelled run is still written
	_ = r.Store.AddLog(context.WithoutCancel(r.runContext()), entry)
	events.NewTestLogEntry.Trigger(events.NewTestLogEntryPayload{TestRunID: entry.TestRunID, Entry: entry})
}

// runContext is the context of the run, or a background context before the
// run began.
func (r *Runner) runContext() context.Context {
	r.ctlMu.Lock()
	defer r.ctlMu.Unlock()
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}

// FailRun logs why the run can't go on and marks it as failed, so callers
// waiting for the result (CLI, MCP, CI) don't take it for a pass.
func (r *Runner) FailRun(format string, params ...interface{}) {
	r.LogError(format, params...)
	if r.ProtocolWriter != nil {
		r.ProtocolWriter.Fail()
	}
}
