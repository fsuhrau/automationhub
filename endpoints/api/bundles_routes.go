package api

import (
	"context"

	bundlesapi "github.com/fsuhrau/automationhub/endpoints/api/bundles"
	"github.com/fsuhrau/automationhub/hub/runs"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/appstore"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

func (s *Service) registerBundleRoutes(appApi *gin.RouterGroup) {
	bundlesapi.New(s.appService()).Register(appApi)
}

// appService builds the app and bundle use cases on the database, the
// bundle storage folder and the run state of this master.
func (s *Service) appService() *appsvc.Service {
	return appsvc.New(appsvc.Deps{
		Repo:     appstore.New(s.db),
		Files:    appsvc.NewOSFiles(apps.AppStoragePath),
		Analyser: appsvc.FileAnalyser{},
		Running:  func(runID uint) bool { return s.runState(runID) == runs.RunRunning },
		Summarize: func(testRuns []models.TestRun) ([]appsvc.RunSummary, error) {
			return s.testService().Summaries(context.Background(), testRuns)
		},
	})
}
