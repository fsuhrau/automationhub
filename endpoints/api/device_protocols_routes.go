package api

import "github.com/gin-gonic/gin"

func (s *Service) registerDeviceProtocolRoutes(projectApi *gin.RouterGroup) {
	s.deviceHandler().RegisterProtocols(projectApi)
}
