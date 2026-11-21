package overview_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/api/overview"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	svc "github.com/fsuhrau/automationhub/service/overview"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// fakeService answers every call with its error and records the project.
type fakeService struct {
	err       error
	projectID uint
	project   *models.Project
}

func (f *fakeService) Get(_ context.Context, projectID uint) (*svc.Overview, error) {
	f.projectID = projectID
	return &svc.Overview{Apps: 2}, f.err
}

func (f *fakeService) Stats(_ context.Context, project *models.Project) (*svc.Stats, error) {
	f.project = project
	return &svc.Stats{AppsCount: 3}, f.err
}

// router serves the handler on the routes of the project 7.
func router(s overview.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	project := r.Group("/project", func(c *gin.Context) {
		c.Set(httpx.ProjectKey, &models.Project{Model: models.Model{ID: 7}})
	})
	h := overview.New(s)
	h.RegisterOverview(project)
	h.RegisterStats(project)
	return r
}

func get(r http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
	return rec
}

func TestOverviewAndStatsAreOfTheProjectOfTheRequest(t *testing.T) {
	f := &fakeService{}
	r := router(f)

	if rec := get(r, "/project/overview"); rec.Code != http.StatusOK || f.projectID != 7 {
		t.Errorf("overview: %d %s, project %d", rec.Code, rec.Body, f.projectID)
	}
	if rec := get(r, "/project/stats"); rec.Code != http.StatusOK || f.project == nil || f.project.ID != 7 {
		t.Errorf("stats: %d %s, project %+v", rec.Code, rec.Body, f.project)
	}
}

func TestOverviewAndStatsShowServiceErrors(t *testing.T) {
	r := router(&fakeService{err: apperr.Internal("db")})

	for _, path := range []string{"/project/overview", "/project/stats"} {
		if rec := get(r, path); rec.Code != http.StatusInternalServerError {
			t.Errorf("%s: %d %s", path, rec.Code, rec.Body)
		}
	}
}
