package runs

import (
	"context"
	"sync"

	"github.com/fsuhrau/automationhub/events"
)

// RunState is the lifecycle state of a test run as seen by this master.
type RunState string

const (
	RunRunning RunState = "running"
	RunDone    RunState = "finished"
	// RunUnknown is a run that wasn't started by this master process (e.g.
	// before a restart); it is no longer executing.
	RunUnknown RunState = "not_running"
)

// RunResult summarizes a finished run.
type RunResult struct {
	Success   bool `json:"success"`
	Succeeded int  `json:"succeeded"`
	Unstable  int  `json:"unstable"`
	Failed    int  `json:"failed"`
}

const maxTrackedResults = 1000

// Tracker records which runs are executing and their results. The test
// run model has no status field; completion is only signalled by the
// TestRunFinished event.
type Tracker struct {
	mu       sync.Mutex
	running  map[uint]chan struct{}
	results  map[uint]RunResult
	order    []uint
	onFinish func(runID uint)
}

func NewTracker(onFinish func(runID uint)) *Tracker {
	return &Tracker{
		running:  make(map[uint]chan struct{}),
		results:  make(map[uint]RunResult),
		onFinish: onFinish,
	}
}

func (t *Tracker) Started(runID uint) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.running[runID]; !ok {
		if _, done := t.results[runID]; !done {
			t.running[runID] = make(chan struct{})
		}
	}
}

// Handle implements the TestRunFinished event handler.
func (t *Tracker) Handle(payload events.TestRunFinishedPayload) {
	t.Finished(payload.TestRunID, RunResult{
		Success:   payload.Success,
		Succeeded: payload.Succeeded,
		Unstable:  payload.Unstable,
		Failed:    payload.Failed,
	})
}

func (t *Tracker) Finished(runID uint, result RunResult) {
	t.mu.Lock()
	if ch, ok := t.running[runID]; ok {
		close(ch)
		delete(t.running, runID)
	}
	if _, ok := t.results[runID]; !ok {
		t.order = append(t.order, runID)
		if len(t.order) > maxTrackedResults {
			delete(t.results, t.order[0])
			t.order = t.order[1:]
		}
	}
	t.results[runID] = result
	t.mu.Unlock()
	if t.onFinish != nil {
		t.onFinish(runID)
	}
}

// Status returns the state of the run and its result once finished.
func (t *Tracker) Status(runID uint) (RunState, *RunResult) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.running[runID]; ok {
		return RunRunning, nil
	}
	if r, ok := t.results[runID]; ok {
		return RunDone, &r
	}
	return RunUnknown, nil
}

// Wait blocks until the run finished or ctx is done.
func (t *Tracker) Wait(ctx context.Context, runID uint) (RunState, *RunResult, error) {
	t.mu.Lock()
	ch, running := t.running[runID]
	t.mu.Unlock()
	if running {
		select {
		case <-ch:
		case <-ctx.Done():
			return RunRunning, nil, ctx.Err()
		}
	}
	state, result := t.Status(runID)
	return state, result, nil
}
