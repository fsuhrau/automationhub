package api

import "github.com/gin-gonic/gin"

func (s *Service) registerTestRoutes(appApi *gin.RouterGroup) {
	appApi.POST("/test", s.WithApp(s.newTest))
	appApi.GET("/test/:test_id", s.WithApp(s.getTest))
	appApi.PUT("/test/:test_id", s.WithApp(s.updateTest))
	appApi.POST("/test/:test_id/run", s.WithApp(s.runTest))
	appApi.GET("/tests", s.WithApp(s.getTests))
}
