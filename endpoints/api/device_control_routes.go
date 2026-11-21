package api

import (
	controlapi "github.com/fsuhrau/automationhub/endpoints/control"
	"github.com/gin-gonic/gin"
)

func (s *Service) registerDeviceControlRoutes(projectApi *gin.RouterGroup) {
	controlapi.NewHandler(s.control, s.resolveControlDevice, s.resolveControlApp, requestOwner, s.nodeUrl).WithAllowedOrigins(s.cfg.Cors).Register(projectApi)
}
