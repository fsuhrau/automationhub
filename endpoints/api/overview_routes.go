package api

import (
	overviewapi "github.com/fsuhrau/automationhub/endpoints/api/overview"
	overviewsvc "github.com/fsuhrau/automationhub/service/overview"
	"github.com/fsuhrau/automationhub/storage/overviewstore"
	"github.com/gin-gonic/gin"
)

// overviewService builds the overview and statistics use cases on the
// database, the run summaries and the device pool.
func (s *Service) overviewService() *overviewsvc.Service {
	return overviewsvc.New(overviewsvc.Deps{
		Repo:       overviewstore.New(s.db),
		Summarizer: s.testService(),
		Pool:       s.devicesManager,
	})
}

func (s *Service) registerOverviewRoutes(projectApi *gin.RouterGroup) {
	overviewapi.New(s.overviewService()).RegisterOverview(projectApi)
}
