package api

import "github.com/gin-gonic/gin"

func (s *Service) registerAppRoutes(projectApi *gin.RouterGroup) {
	projectApi.POST("/app", s.WithProject(s.createApp))
	projectApi.GET("/apps", s.WithProject(s.getApps))
}

// registerAppDetailRoutes adds the routes of one app: the app itself and
// its parameters.
func (s *Service) registerAppDetailRoutes(appApi *gin.RouterGroup) {
	appApi.PUT("/", s.WithApp(s.updateApp))
	appApi.POST("/parameter", s.WithApp(s.addAppParameter))
	appApi.PUT("/parameter/:parameter_id", s.WithApp(s.updateAppParameter))
	appApi.DELETE("/parameter/:parameter_id", s.WithApp(s.removeAppParameter))
}
