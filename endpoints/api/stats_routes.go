package api

import "github.com/gin-gonic/gin"

func (s *Service) registerStatsRoutes(api *gin.RouterGroup) {
	api.GET("/stats", s.getStats)
}
