package api

import "github.com/gin-gonic/gin"

func (s *Service) registerBaselineRoutes(appApi *gin.RouterGroup) {
	appApi.GET("/test/:test_id/run/:run_id/:protocol_id/baselines", s.WithApp(s.getProtocolBaselines))
	appApi.GET("/test/:test_id/run/:run_id/:protocol_id/baseline/:baseline_run_id", s.WithApp(s.getProtocolBaseline))
}
