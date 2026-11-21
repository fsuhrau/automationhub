package api

import (
	testsapi "github.com/fsuhrau/automationhub/endpoints/api/tests"
	"github.com/gin-gonic/gin"
)

func (s *Service) registerTestRoutes(appApi *gin.RouterGroup) {
	testsapi.New(s.testService()).Register(appApi)
	appApi.POST("/test/:test_id/run", s.WithApp(s.runTest))
}
