package api

import (
	overviewapi "github.com/fsuhrau/automationhub/endpoints/api/overview"
	"github.com/gin-gonic/gin"
)

// registerStatsRoutes adds the statistics of a project. The group of the
// project is made here, since the stats routes register before the project
// group of RegisterRoutes exists.
func (s *Service) registerStatsRoutes(api *gin.RouterGroup) {
	overviewapi.New(s.overviewService()).RegisterStats(api.Group("/:project_id", s.ResolveProject))
}
