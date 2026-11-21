package api

import "github.com/gin-gonic/gin"

func (s *Service) registerDeviceRoutes(projectApi *gin.RouterGroup) {
	projectApi.GET("/device/:device_id", s.WithProject(s.getDevice))
	projectApi.DELETE("/device/:device_id", s.WithProject(s.deleteDevice))
	projectApi.POST("/device/:device_id/unlock", s.WithProject(s.unlockDevice))
	projectApi.PUT("/device/:device_id", s.WithProject(s.updateDevice))
	projectApi.POST("/device/:device_id/tests", s.WithProject(s.deviceRunTests))
	projectApi.GET("/devices", s.WithProject(s.getDevices))
}
