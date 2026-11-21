package runs_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/api/runs"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/service/run"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// fakeService answers every call with its fields and records the requests.
type fakeService struct {
	err         error
	started     *models.TestRun
	startedWith run.StartRequest
	cancel      [2]uint
	key         run.Key
	ref         testsvc.TestRef
	list        run.ListQuery
}

func (f *fakeService) Start(_ context.Context, req run.StartRequest) (run.RunRef, error) {
	f.startedWith = req
	return run.RunRef{Run: f.started}, f.err
}

func (f *fakeService) Cancel(_ context.Context, projectID, runID uint) error {
	f.cancel = [2]uint{projectID, runID}
	return f.err
}

func (f *fakeService) Get(_ context.Context, key run.Key) (*run.Page, error) {
	f.key = key
	return &run.Page{PrevRunID: 1}, f.err
}

func (f *fakeService) Last(_ context.Context, ref testsvc.TestRef) (*run.Page, error) {
	f.ref = ref
	return &run.Page{NextRunID: 2}, f.err
}

func (f *fakeService) List(_ context.Context, q run.ListQuery) (*run.RunsPage, error) {
	f.list = q
	return &run.RunsPage{Runs: []run.ListEntry{}, More: true}, f.err
}

// router serves the handler on the routes of an app of the project 7, the
// way the api resolves them.
func router(svc runs.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	app := r.Group("/app", func(c *gin.Context) {
		c.Set(httpx.ProjectKey, &models.Project{Model: models.Model{ID: 7}})
		c.Set(httpx.AppKey, &models.App{Model: models.Model{ID: 8}})
	})
	runs.New(svc).Register(app)
	return r
}

func do(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec
}

func TestStartAnswersWithTheRunAndPassesTheRequestOn(t *testing.T) {
	f := &fakeService{started: &models.TestRun{Model: models.Model{ID: 5}, TestID: 3}}

	rec := do(router(f), "POST", "/app/test/3/run", `{"appBinaryId":4,"startUrl":"https://x.test","params":"a=1"}`)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"id":5`) {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	req := f.startedWith
	if req.Project.ID != 7 || req.TestID != 3 || req.BundleID != 4 || req.StartURL != "https://x.test" || req.Params != "a=1" {
		t.Fatalf("request %+v", req)
	}
}

func TestStartRefusals(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		body    string
		svcErr  error
		status  int
		message string
	}{
		{"invalid test id", "/app/test/abc/run", `{}`, nil, 400, "invalid test id"},
		{"malformed body", "/app/test/3/run", `{"appBinaryId":"x"}`, nil, 400, ""},
		{"service error", "/app/test/3/run", `{}`, apperr.Conflict("no device available"), 409, "no device available"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeService{err: tc.svcErr}

			rec := do(router(f), "POST", tc.path, tc.body)

			if rec.Code != tc.status || tc.message != "" && !strings.Contains(rec.Body.String(), tc.message) {
				t.Fatalf("status %d: %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestCancelAnswersCancelled(t *testing.T) {
	f := &fakeService{}

	rec := do(router(f), "POST", "/app/test/3/run/9/cancel", "")

	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"status":"cancelled"}` {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if f.cancel != [2]uint{7, 9} {
		t.Fatalf("cancelled %v, want project 7 run 9", f.cancel)
	}
}

func TestCancelRefusals(t *testing.T) {
	invalid := do(router(&fakeService{}), "POST", "/app/test/3/run/abc/cancel", "")
	if invalid.Code != http.StatusBadRequest || !strings.Contains(invalid.Body.String(), "invalid run id") {
		t.Errorf("invalid run id: %d %s", invalid.Code, invalid.Body)
	}
	missing := do(router(&fakeService{err: apperr.NotFound("not found: no active runner")}), "POST", "/app/test/3/run/9/cancel", "")
	if missing.Code != http.StatusNotFound {
		t.Errorf("not running: %d %s", missing.Code, missing.Body)
	}
}

func TestReadsPassTheScopeAndIDsOfThePathOn(t *testing.T) {
	f := &fakeService{}
	r := router(f)

	if rec := do(r, "GET", "/app/test/3/run/9", ""); rec.Code != http.StatusOK {
		t.Fatalf("get: %d %s", rec.Code, rec.Body)
	}
	want := run.Key{TestRef: testsvc.TestRef{Scope: testsvc.Scope{ProjectID: 7, AppID: 8}, TestID: 3}, RunID: 9}
	if f.key != want {
		t.Errorf("key %+v, want %+v", f.key, want)
	}
	if rec := do(r, "GET", "/app/test/3/runs/last", ""); rec.Code != http.StatusOK || f.ref != want.TestRef {
		t.Errorf("last: %d %s, ref %+v", rec.Code, rec.Body, f.ref)
	}
}

func TestListReadsTheLimitAndTheCursor(t *testing.T) {
	tests := []struct {
		query  string
		before uint
		limit  int
	}{
		{"", 0, run.PageDefault},
		{"?limit=2&before=40", 40, 2},
		{"?limit=5000", 0, run.PageMax},
		{"?limit=-1&before=-4", 0, run.PageDefault},
		{"?before=abc", 0, run.PageDefault},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			f := &fakeService{}

			rec := do(router(f), "GET", "/app/test/3/runs"+tc.query, "")

			if rec.Code != http.StatusOK || f.list.Before != tc.before || f.list.Limit != tc.limit || f.list.TestID != 3 {
				t.Fatalf("status %d, query %+v", rec.Code, f.list)
			}
		})
	}
}

func TestReadsRefuseNonNumericIDsAndShowServiceErrors(t *testing.T) {
	r := router(&fakeService{})
	for _, path := range []string{"/app/test/x/runs", "/app/test/x/runs/last", "/app/test/x/run/1", "/app/test/3/run/x"} {
		if rec := do(r, "GET", path, ""); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", path, rec.Code, rec.Body)
		}
	}
	failing := router(&fakeService{err: apperr.NotFound("test 3 not found")})
	for _, path := range []string{"/app/test/3/runs", "/app/test/3/runs/last", "/app/test/3/run/9"} {
		if rec := do(failing, "GET", path, ""); rec.Code != http.StatusNotFound {
			t.Errorf("%s: %d %s", path, rec.Code, rec.Body)
		}
	}
}
