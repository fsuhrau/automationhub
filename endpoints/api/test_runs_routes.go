package api

import (
	runsapi "github.com/fsuhrau/automationhub/endpoints/api/runs"
	"github.com/gin-gonic/gin"
)

func (s *Service) registerTestRunRoutes(appApi *gin.RouterGroup) {
	runsapi.New(s.runService()).Register(appApi)
}
