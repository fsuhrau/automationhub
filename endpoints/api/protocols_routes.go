package api

import "github.com/gin-gonic/gin"

func (s *Service) registerProtocolRoutes(appApi *gin.RouterGroup) {
	appApi.GET("/test/:test_id/run/:run_id/:protocol_id", s.WithApp(s.getTestRunProtocol))
}
