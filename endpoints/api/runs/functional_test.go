package runs_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/endpoints/api/internal/runfixture"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

// listWorld is the world with three more runs of the own test that ran with
// a bundle; the newest has two ended protocols.
type listWorld struct {
	*runfixture.World
	ids      []uint
	laterEnd time.Time
}

func newListWorld(t *testing.T) *listWorld {
	t.Helper()
	w := &listWorld{World: runfixture.New(t)}
	bundle := models.AppBinary{AppID: w.MineApp.ID, Name: "Game", Version: "1.2"}
	w.DB.Create(&bundle)
	for i := 0; i < 3; i++ {
		r := testkit.TestRun(w.DB, func(r *models.TestRun) {
			r.TestID, r.AppBinaryID, r.Parameter = w.Mine.ID, bundle.ID, fmt.Sprintf("Run=%d", i)
		})
		w.ids = append(w.ids, r.ID)
	}
	end := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	w.laterEnd = end.Add(time.Minute)
	w.protocol(w.ids[2], models.TestResultSuccess, &end)
	w.protocol(w.ids[2], models.TestResultFailed, &w.laterEnd)
	return w
}

func (w *listWorld) protocol(runID uint, result models.TestResultState, end *time.Time) {
	testkit.Protocol(w.DB, func(p *models.TestProtocol) { p.TestRunID, p.TestResult, p.EndedAt = runID, result, end })
}

func (w *listWorld) list(t *testing.T, query string) run.RunsPage {
	t.Helper()
	return testkit.Decode[run.RunsPage](t, w.Do("GET", w.Path("/test/%d/runs%s", w.Mine.ID, query), nil))
}

// The test page pages through the runs of a test, newest first, with what
// each ran with and when it ended.
func TestTheRunsPageListsRunsNewestFirstWithBundleAndEnd(t *testing.T) {
	w := newListWorld(t)

	page := w.list(t, "?limit=2")

	if len(page.Runs) != 2 || !page.More {
		t.Fatalf("first page %+v", page)
	}
	first := page.Runs[0]
	if first.ID != w.ids[2] || first.Status != "failed" || first.Passed != 1 || first.Failed != 1 {
		t.Errorf("newest run %+v", first.RunSummary)
	}
	if first.Bundle == nil || first.Bundle.Version != "1.2" || first.Parameter != "Run=2" {
		t.Errorf("newest run ran with %+v %q", first.Bundle, first.Parameter)
	}
	if first.EndedAt == nil || !first.EndedAt.Equal(w.laterEnd) {
		t.Errorf("ended at %v, want the last protocol's end %v", first.EndedAt, w.laterEnd)
	}
	if page.Runs[1].EndedAt != nil || page.Runs[1].Status != "empty" {
		t.Errorf("a run without protocols: %+v", page.Runs[1])
	}
}

func TestTheRunsPageContinuesBeforeTheLastRunOfAPage(t *testing.T) {
	w := newListWorld(t)

	older := w.list(t, fmt.Sprintf("?limit=2&before=%d", w.ids[1]))
	if len(older.Runs) != 2 || older.Runs[0].ID != w.ids[0] || older.Runs[1].ID != w.Run2.ID || !older.More {
		t.Errorf("second page %+v", older)
	}
	if last := w.list(t, fmt.Sprintf("?limit=2&before=%d", w.Run2.ID)); len(last.Runs) != 1 || last.Runs[0].ID != w.Run1.ID || last.More {
		t.Errorf("last page %+v", last)
	}
	if all := w.list(t, ""); len(all.Runs) != 5 || all.More {
		t.Errorf("default page: %d runs, more %v", len(all.Runs), all.More)
	}
}

// A run that cannot start or be cancelled says why, and nothing starts in
// another project.
func TestStartAndCancelRefusalsThroughTheRouter(t *testing.T) {
	w := runfixture.New(t)
	tests := []struct {
		name    string
		path    string
		body    any
		status  int
		message string
	}{
		{"unknown test", w.Path("/test/99/run"), map[string]any{"appBinaryId": 1}, http.StatusNotFound, "not found: test 99"},
		{"test of another project", w.Path("/test/%d/run", w.Foreign.ID), map[string]any{}, http.StatusNotFound, fmt.Sprintf("not found: test %d", w.Foreign.ID)},
		{"invalid test id", w.Path("/test/abc/run"), map[string]any{}, http.StatusBadRequest, "invalid test id"},
		{"malformed body", w.Path("/test/%d/run", w.Mine.ID), map[string]any{"appBinaryId": "x"}, http.StatusBadRequest, ""},
		{"unknown bundle", w.Path("/test/%d/run", w.Mine.ID), map[string]any{"appBinaryId": 99}, http.StatusNotFound,
			fmt.Sprintf("not found: app bundle 99 for app %d", w.MineApp.ID)},
		{"run that is not running", w.Path("/test/%d/run/%d/cancel", w.Mine.ID, w.Run1.ID), nil, http.StatusNotFound,
			fmt.Sprintf("not found: no active runner for test run %d", w.Run1.ID)},
		{"run of another project", w.Path("/test/%d/run/%d/cancel", w.Foreign.ID, w.ForeignRun.ID), nil, http.StatusNotFound,
			fmt.Sprintf("not found: test run %d", w.ForeignRun.ID)},
		{"invalid run id", w.Path("/test/%d/run/abc/cancel", w.Mine.ID), nil, http.StatusBadRequest, "invalid run id"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := w.Do("POST", tc.path, tc.body)

			if rec.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if got := testkit.Decode[httpx.ErrorResponse](t, rec).Message; tc.message != "" && got != tc.message {
				t.Fatalf("message %q, want %q", got, tc.message)
			}
		})
	}
}
