package test

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/hub/runs"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
)

// RunSummary is one run in a list: when it ran and how it went. The status
// is running, passed, failed, unstable, interrupted, not_run or empty; the
// same rules as the run page. The app service shows the same summaries on the
// bundle page, so the type lives there.
type RunSummary = appsvc.RunSummary

type counts struct{ passed, failed, unstable, open, logErrors int }

// summarize is the run status of the run page: a running run is running;
// otherwise failures, then unstable tests, then tests that never finished
// decide; a run without results did not run when its log has errors.
func summarize(state runs.RunState, c counts) string {
	total := c.passed + c.failed + c.unstable + c.open
	switch {
	case state == runs.RunRunning:
		return "running"
	case total == 0 && c.logErrors > 0:
		return "not_run"
	case total == 0:
		return "empty"
	case c.failed > 0:
		return "failed"
	case c.unstable > 0:
		return "unstable"
	case c.open > 0:
		return "interrupted"
	default:
		return "passed"
	}
}

// Summaries gives each run its status and result counts, in the order of
// runs. It counts the tests themselves; a runner protocol wraps one.
func (s *Service) Summaries(ctx context.Context, testRuns []models.TestRun) ([]RunSummary, error) {
	summaries := make([]RunSummary, len(testRuns))
	if len(testRuns) == 0 {
		return summaries, nil
	}
	byRun, err := s.countRuns(ctx, runIDs(testRuns))
	if err != nil {
		return nil, err
	}
	for i, r := range testRuns {
		n := byRun[r.ID]
		state := s.runState(r.ID)
		summaries[i] = RunSummary{
			ID: r.ID, CreatedAt: r.CreatedAt, State: state,
			Passed: n.passed, Failed: n.failed, Unstable: n.unstable, Open: n.open,
			Status: summarize(state, n),
		}
	}
	return summaries, nil
}

func runIDs(testRuns []models.TestRun) []uint {
	ids := make([]uint, len(testRuns))
	for i, r := range testRuns {
		ids[i] = r.ID
	}
	return ids
}

// countRuns reads the result and log error counts of the runs.
func (s *Service) countRuns(ctx context.Context, ids []uint) (map[uint]counts, error) {
	results, err := s.repo.RunResults(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("count run results: %w", err)
	}
	logErrors, err := s.repo.RunLogErrors(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("count run log errors: %w", err)
	}
	byRun := map[uint]counts{}
	for _, r := range results {
		n := byRun[r.TestRunID]
		addResult(&n, r)
		byRun[r.TestRunID] = n
	}
	for _, e := range logErrors {
		n := byRun[e.TestRunID]
		n.logErrors = e.N
		byRun[e.TestRunID] = n
	}
	return byRun, nil
}

func addResult(n *counts, r ResultCount) {
	switch r.TestResult {
	case models.TestResultSuccess:
		n.passed += r.N
	case models.TestResultFailed:
		n.failed += r.N
	case models.TestResultUnstable:
		n.unstable += r.N
	default:
		n.open += r.N
	}
}
