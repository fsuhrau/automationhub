package run_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestReportHasTheRunAndTheNewestLogOldestFirst(t *testing.T) {
	e := newEnv()
	e.repo.reportRun = &models.TestRun{Model: models.Model{ID: 7}}
	e.repo.logTail = []models.TestRunLogEntry{{Log: "newest"}, {Log: "older"}}

	got, err := e.svc.Report(bg(), 1, 7)

	if err != nil || got.Run.ID != 7 || len(got.Log) != 2 || got.Log[0].Log != "older" || got.Log[1].Log != "newest" {
		t.Fatalf("report %+v, err %v", got, err)
	}
	if e.repo.tailLimit != run.ReportLogLines {
		t.Errorf("asked for %d log lines, want %d", e.repo.tailLimit, run.ReportLogLines)
	}
}

func TestReportFailsWhereTheRepoFails(t *testing.T) {
	for _, method := range []string{"RunInProject", "ReportRun", "LogTail"} {
		t.Run(method, func(t *testing.T) {
			e := newEnv()
			e.repo.fail[method] = apperr.NotFound(method)

			_, err := e.svc.Report(bg(), 1, 7)

			assertKind(t, err, apperr.KindNotFound)
		})
	}
}

func TestProgressCountsTheEndedProtocols(t *testing.T) {
	e := newEnv()
	e.repo.progress = run.Progress{Total: 4, Ended: 3}

	got, err := e.svc.Progress(bg(), 1, 7)

	if err != nil || got != e.repo.progress {
		t.Fatalf("progress %+v, err %v", got, err)
	}
}

func TestProgressOfARunOfAnotherProjectIsNotFound(t *testing.T) {
	e := newEnv()
	e.repo.fail["RunInProject"] = apperr.NotFound("run")

	_, err := e.svc.Progress(bg(), 2, 7)

	assertKind(t, err, apperr.KindNotFound)
}
