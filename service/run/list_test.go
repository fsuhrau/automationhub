package run_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

func runs3() []models.TestRun {
	bundle := &models.AppBinary{Model: models.Model{ID: 9}, Name: "Game", Version: "1.2"}
	return []models.TestRun{
		{Model: models.Model{ID: 30}, Parameter: "Run=3", AppBinary: bundle},
		{Model: models.Model{ID: 20}, Parameter: "Run=2"},
		{Model: models.Model{ID: 10}, Parameter: "Run=1"},
	}
}

func listQuery(limit int) run.ListQuery {
	return run.ListQuery{TestRef: ref(), Before: 40, Limit: limit}
}

func TestListPagesTheRunsWithTheirBundleAndEnd(t *testing.T) {
	e := newEnv()
	e.repo.testRuns = runs3()
	e.repo.ends = []run.RunEnd{{TestRunID: 30, EndedAt: at(100)}, {TestRunID: 30, EndedAt: at(300)}, {TestRunID: 30, EndedAt: at(200)}}

	page, err := e.svc.List(bg(), listQuery(2))

	if err != nil {
		t.Fatal(err)
	}
	if !page.More || len(page.Runs) != 2 {
		t.Fatalf("page %+v, want 2 runs and more", page)
	}
	if e.repo.limit != 3 || e.repo.before != 40 {
		t.Errorf("repo asked for %d runs before %d, want one more than the limit", e.repo.limit, e.repo.before)
	}
	first, second := page.Runs[0], page.Runs[1]
	if first.ID != 30 || first.Status != "passed" || first.Parameter != "Run=3" || first.Bundle == nil || first.Bundle.Version != "1.2" {
		t.Errorf("newest run %+v", first)
	}
	if first.EndedAt == nil || !first.EndedAt.Equal(at(300)) {
		t.Errorf("ended at %v, want the latest protocol end", first.EndedAt)
	}
	if second.EndedAt != nil || second.Bundle != nil {
		t.Errorf("run without protocols or bundle %+v", second)
	}
	if len(e.repo.runIDsSeen) != 2 {
		t.Errorf("ends asked for runs %v, want the 2 listed", e.repo.runIDsSeen)
	}
}

func TestListOfTheLastPageHasNoMore(t *testing.T) {
	e := newEnv()
	e.repo.testRuns = runs3()

	page, err := e.svc.List(bg(), listQuery(5))

	if err != nil || page.More || len(page.Runs) != 3 {
		t.Fatalf("page %+v, err %v", page, err)
	}
}

func TestListWithoutRunsDoesNotAskForEnds(t *testing.T) {
	e := newEnv()
	e.repo.fail["RunEnds"] = apperr.Internal("must not be asked")

	page, err := e.svc.List(bg(), listQuery(5))

	if err != nil || page.More || len(page.Runs) != 0 || page.Runs == nil {
		t.Fatalf("page %+v, err %v", page, err)
	}
}

func TestListErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*env)
		want  apperr.Kind
	}{
		{"test of another project", func(e *env) { e.repo.fail["TestInScope"] = apperr.NotFound("test") }, apperr.KindNotFound},
		{"runs", func(e *env) { e.repo.fail["TestRuns"] = apperr.Internal("db") }, apperr.KindInternal},
		{"summaries", func(e *env) { e.summary.err = apperr.Internal("db") }, apperr.KindInternal},
		{"ends", func(e *env) { e.repo.fail["RunEnds"] = apperr.Internal("db") }, apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv()
			e.repo.testRuns = runs3()
			tc.setup(e)

			page, err := e.svc.List(bg(), listQuery(5))

			assertKind(t, err, tc.want)
			if page != nil {
				t.Fatalf("page %+v with an error", page)
			}
		})
	}
}
