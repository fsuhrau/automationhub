package protocols_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/api/protocols"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/service/protocol"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// fakeService answers every call with its error and records the requests.
type fakeService struct {
	err    error
	key    protocol.Key
	filter protocol.Filter
	choice protocol.Choice
}

func (f *fakeService) Page(_ context.Context, key protocol.Key) (*models.TestRun, error) {
	f.key = key
	return &models.TestRun{TestID: 3}, f.err
}

func (f *fakeService) Baselines(_ context.Context, key protocol.Key, filter protocol.Filter) (*protocol.Baselines, error) {
	f.key, f.filter = key, filter
	return &protocol.Baselines{Candidates: []protocol.Candidate{}, Tags: []string{}}, f.err
}

func (f *fakeService) Baseline(_ context.Context, key protocol.Key, choice protocol.Choice) (*protocol.Baseline, error) {
	f.key, f.choice = key, choice
	return &protocol.Baseline{}, f.err
}

// router serves the handler on the routes of an app of the project 7.
func router(svc protocols.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	app := r.Group("/app", func(c *gin.Context) {
		c.Set(httpx.ProjectKey, &models.Project{Model: models.Model{ID: 7}})
		c.Set(httpx.AppKey, &models.App{Model: models.Model{ID: 8}})
	})
	h := protocols.New(svc)
	h.RegisterProtocols(app)
	h.RegisterBaselines(app)
	return r
}

func get(r http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
	return rec
}

func wantKey() protocol.Key {
	return protocol.Key{TestRef: testsvc.TestRef{Scope: testsvc.Scope{ProjectID: 7, AppID: 8}, TestID: 3}, RunID: 4, ProtocolID: 5}
}

func TestPageAndBaselinesPassTheScopeAndIDsOfThePathOn(t *testing.T) {
	for _, path := range []string{"/app/test/3/run/4/5", "/app/test/3/run/4/5/baselines", "/app/test/3/run/4/5/baseline/6"} {
		f := &fakeService{}

		rec := get(router(f), path)

		if rec.Code != http.StatusOK || f.key != wantKey() {
			t.Errorf("%s: status %d, key %+v, want %+v", path, rec.Code, f.key, wantKey())
		}
	}
}

func TestBaselinesReadTheFilters(t *testing.T) {
	tests := []struct {
		query string
		want  protocol.Filter
	}{
		{"", protocol.Filter{Limit: protocol.DefaultCandidates}},
		{"?tag=release&q=1.1&anyDevice=true&limit=7", protocol.Filter{Tag: "release", Query: "1.1", AnyDevice: true, Limit: 7}},
		{"?anyDevice=yes&limit=9999", protocol.Filter{Limit: protocol.MaxCandidates}},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			f := &fakeService{}

			rec := get(router(f), "/app/test/3/run/4/5/baselines"+tc.query)

			if rec.Code != http.StatusOK || f.filter != tc.want {
				t.Fatalf("status %d, filter %+v, want %+v", rec.Code, f.filter, tc.want)
			}
		})
	}
}

func TestBaselineReadsTheRunAndTheOptionalDevice(t *testing.T) {
	tests := []struct {
		query  string
		device *uint
	}{
		{"", nil},
		{"?deviceId=abc", nil},
		{"?deviceId=12", new(uint)},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			f := &fakeService{}

			rec := get(router(f), "/app/test/3/run/4/5/baseline/6"+tc.query)

			if rec.Code != http.StatusOK || f.choice.RunID != 6 {
				t.Fatalf("status %d, run %d", rec.Code, f.choice.RunID)
			}
			if (f.choice.DeviceID == nil) != (tc.device == nil) || f.choice.DeviceID != nil && *f.choice.DeviceID != 12 {
				t.Fatalf("device %v, want one only for a number", f.choice.DeviceID)
			}
		})
	}
}

func TestRefusalsAndServiceErrors(t *testing.T) {
	paths := []string{
		"/app/test/3/run/4/5", "/app/test/3/run/4/5/baselines", "/app/test/3/run/4/5/baseline/6",
	}
	for _, path := range paths {
		failing := get(router(&fakeService{err: apperr.NotFound("protocol 5 not found in run 4")}), path)
		if failing.Code != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404", path, failing.Code)
		}
	}
	for _, path := range []string{
		"/app/test/x/run/4/5", "/app/test/3/run/x/5", "/app/test/3/run/4/x", "/app/test/3/run/4/x/baselines",
		"/app/test/3/run/4/5/baseline/x", "/app/test/3/run/4/x/baseline/6",
	} {
		if rec := get(router(&fakeService{}), path); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", path, rec.Code)
		}
	}
}
