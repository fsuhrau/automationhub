package teststore_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestRecentRunsAreNewestFirstAndLimited(t *testing.T) {
	w := newWorld(t)
	var ids []uint
	for i := 0; i < 3; i++ {
		ids = append(ids, testkit.TestRun(w.db, func(r *models.TestRun) { r.TestID = w.mine.ID }).ID)
	}
	testkit.TestRun(w.db, func(r *models.TestRun) { r.TestID = w.foreign.ID })

	got, err := w.store.RecentRuns(bg(), w.mine.ID, 2)

	if err != nil || len(got) != 2 || got[0].ID != ids[2] || got[1].ID != ids[1] {
		t.Fatalf("recent runs %+v, err %v; want %d then %d", got, err, ids[2], ids[1])
	}
}

func TestRunResultsCountLeafProtocolsPerResult(t *testing.T) {
	w := newWorld(t)
	run := testkit.TestRun(w.db, func(r *models.TestRun) { r.TestID = w.mine.ID })
	runner := testkit.Protocol(w.db, func(p *models.TestProtocol) { p.TestRunID = run.ID })
	for _, res := range []models.TestResultState{models.TestResultSuccess, models.TestResultSuccess, models.TestResultFailed} {
		testkit.Protocol(w.db, func(p *models.TestProtocol) {
			p.TestRunID, p.ParentTestProtocolID, p.TestResult = run.ID, &runner.ID, res
		})
	}

	got, err := w.store.RunResults(bg(), []uint{run.ID})

	counts := map[models.TestResultState]int{}
	for _, r := range got {
		counts[r.TestResult] += r.N
	}
	if err != nil || counts[models.TestResultSuccess] != 2 || counts[models.TestResultFailed] != 1 || len(counts) != 2 {
		t.Fatalf("counts %v, err %v; the wrapping runner must not count", counts, err)
	}
}

func TestRunLogErrorsCountOnlyErrors(t *testing.T) {
	w := newWorld(t)
	run := testkit.TestRun(w.db, func(r *models.TestRun) { r.TestID = w.mine.ID })
	w.db.Create(&models.TestRunLogEntry{TestRunID: run.ID, Level: "error", Log: "a"})
	w.db.Create(&models.TestRunLogEntry{TestRunID: run.ID, Level: "error", Log: "b"})
	w.db.Create(&models.TestRunLogEntry{TestRunID: run.ID, Level: "info", Log: "c"})

	got, err := w.store.RunLogErrors(bg(), []uint{run.ID})

	if err != nil || len(got) != 1 || got[0].TestRunID != run.ID || got[0].N != 2 {
		t.Fatalf("log errors %+v, err %v", got, err)
	}
}

func TestRunQueriesMapDatabaseErrors(t *testing.T) {
	queries := map[string]struct {
		table string
		call  func(w *world) error
	}{
		"recent runs": {"test_runs", func(w *world) error { _, err := w.store.RecentRuns(bg(), 1, 1); return err }},
		"results":     {"test_protocols", func(w *world) error { _, err := w.store.RunResults(bg(), []uint{1}); return err }},
		"log errors":  {"test_run_log_entries", func(w *world) error { _, err := w.store.RunLogErrors(bg(), []uint{1}); return err }},
	}
	for name, q := range queries {
		t.Run(name, func(t *testing.T) {
			w := newWorld(t)
			w.drop(t, q.table)

			wantKind(t, q.call(w), apperr.KindInternal)
		})
	}
}
