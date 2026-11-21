package api

import "github.com/gin-gonic/gin"

func (s *Service) registerOverviewRoutes(projectApi *gin.RouterGroup) {
	projectApi.GET("/overview", s.WithProject(s.getOverview))
}
