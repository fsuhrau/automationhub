package api

import "github.com/gin-gonic/gin"

func (s *Service) registerDeviceProtocolRoutes(projectApi *gin.RouterGroup) {
	projectApi.GET("/device/:device_id/protocols", s.WithProject(s.getDeviceProtocols))
}
