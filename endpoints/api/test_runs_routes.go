package api

import "github.com/gin-gonic/gin"

func (s *Service) registerTestRunRoutes(appApi *gin.RouterGroup) {
	appApi.GET("/test/:test_id/runs", s.WithApp(s.getTestRuns))
	appApi.GET("/test/:test_id/runs/last", s.WithApp(s.getLastTestRun))
	appApi.GET("/test/:test_id/run/:run_id", s.WithApp(s.getTestRun))
	appApi.POST("/test/:test_id/run/:run_id/cancel", s.WithApp(s.cancelTestRun))
}
