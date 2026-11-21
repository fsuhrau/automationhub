package api

import "github.com/gin-gonic/gin"

func (s *Service) registerDataRoutes(api *gin.RouterGroup) {
	api.GET("/data/:name", s.getData)
}
