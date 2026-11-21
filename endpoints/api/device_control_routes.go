package api

import (
	controlapi "github.com/fsuhrau/automationhub/endpoints/control"
	"github.com/gin-gonic/gin"
)

func (s *Service) registerDeviceControlRoutes(projectApi *gin.RouterGroup) {
	devices := s.deviceHandler()
	controlapi.NewHandler(s.control, devices.ResolveControlDevice, devices.ResolveControlApp, requestOwner, s.nodeUrl).WithAllowedOrigins(s.cfg.Cors).Register(projectApi)
}
