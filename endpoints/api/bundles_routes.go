package api

import (
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
		Repo:      appstore.New(s.db),
		Files:     appsvc.NewOSFiles(apps.AppStoragePath),
		Analyser:  appsvc.FileAnalyser{},
		Running:   func(runID uint) bool { return s.runState(runID) == runs.RunRunning },
		Summarize: s.summarizeBundleRuns,
	})
}

// summarizeBundleRuns is summarizeRuns in the shape the app service uses.
func (s *Service) summarizeBundleRuns(testRuns []models.TestRun) ([]appsvc.RunSummary, error) {
	summaries, err := s.summarizeRuns(testRuns)
	if err != nil {
		return nil, err
	}
	out := make([]appsvc.RunSummary, len(summaries))
	for i := range summaries {
		out[i] = appsvc.RunSummary(summaries[i])
	}
	return out, nil
}
