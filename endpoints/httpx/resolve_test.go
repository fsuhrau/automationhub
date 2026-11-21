package httpx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

type fakeProjects struct {
	err   error
	asked string
}

func (f *fakeProjects) Get(_ context.Context, identifier string) (*models.Project, error) {
	f.asked = identifier
	if f.err != nil {
		return nil, f.err
	}
	return &models.Project{Model: models.Model{ID: 4}, Identifier: identifier}, nil
}

type fakeApps struct {
	err   error
	asked appsvc.Scope
}

func (f *fakeApps) App(_ context.Context, sc appsvc.Scope) (*models.App, error) {
	f.asked = sc
	if f.err != nil {
		return nil, f.err
	}
	return &models.App{Model: models.Model{ID: sc.AppID}}, nil
}

func TestResolveProjectStoresTheProject(t *testing.T) {
	projects := &fakeProjects{}
	var got *models.Project
	handler := func(c *gin.Context) { got = httpx.Project(c) }

	rec := serve(http.MethodGet, "/:project_id", "/alpha", handler, httpx.ResolveProject(projects))

	if rec.Code != http.StatusOK || got == nil || got.Identifier != "alpha" || projects.asked != "alpha" {
		t.Fatalf("status %d, project %+v", rec.Code, got)
	}
}

func TestResolveProjectFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
		body string
	}{
		{"unknown project keeps the stable text", apperr.NotFound("project alpha not found"), http.StatusNotFound, `{"message":"record not found"}`},
		{"other failures are reported as they are", apperr.Busy("db is busy"), http.StatusConflict, `{"message":"db is busy"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			handler := func(*gin.Context) { called = true }

			rec := serve(http.MethodGet, "/:project_id", "/alpha", handler, httpx.ResolveProject(&fakeProjects{err: tc.err}))

			if rec.Code != tc.code || rec.Body.String() != tc.body || called {
				t.Fatalf("status %d, body %s, handler called %v", rec.Code, rec.Body, called)
			}
		})
	}
}

func TestResolveAppStoresTheAppOfTheProject(t *testing.T) {
	apps := &fakeApps{}
	var got *models.App
	handler := func(c *gin.Context) { got = httpx.App(c) }
	setProject := func(c *gin.Context) { c.Set(httpx.ProjectKey, &models.Project{Model: models.Model{ID: 4}}) }

	rec := serve(http.MethodGet, "/:app_id", "/9", handler, setProject, httpx.ResolveApp(apps))

	if rec.Code != http.StatusOK || got == nil || got.ID != 9 || apps.asked != (appsvc.Scope{ProjectID: 4, AppID: 9}) {
		t.Fatalf("status %d, app %+v, asked %+v", rec.Code, got, apps.asked)
	}
}

func TestResolveAppFailures(t *testing.T) {
	tests := []struct {
		name string
		path string
		err  error
		code int
		body string
	}{
		{"unknown app", "/9", apperr.NotFound("app 9 not found"), http.StatusNotFound, `{"message":"record not found"}`},
		{"app id is no number", "/abc", nil, http.StatusNotFound, `{"message":"record not found"}`},
		{"other failures", "/9", apperr.Internal("db down"), http.StatusInternalServerError, `{"message":"db down"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setProject := func(c *gin.Context) { c.Set(httpx.ProjectKey, &models.Project{}) }

			rec := serve(http.MethodGet, "/:app_id", tc.path, func(*gin.Context) {}, setProject, httpx.ResolveApp(&fakeApps{err: tc.err}))

			if rec.Code != tc.code || rec.Body.String() != tc.body {
				t.Fatalf("status %d, body %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestResolveKeepsWorkingWhenTheClientIsGone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cancelled := &cancelAware{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(httpx.ResolveProject(cancelled))
	r.GET("/:project_id", func(*gin.Context) {})
	req := httptest.NewRequest(http.MethodGet, "/alpha", nil).WithContext(ctx)

	r.ServeHTTP(httptest.NewRecorder(), req)

	if cancelled.err != nil {
		t.Fatalf("the lookup saw the cancellation: %v", cancelled.err)
	}
}

type cancelAware struct{ err error }

func (f *cancelAware) Get(ctx context.Context, identifier string) (*models.Project, error) {
	f.err = ctx.Err()
	return &models.Project{Identifier: identifier}, nil
}
