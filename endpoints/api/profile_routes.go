package api

import (
	profileapi "github.com/fsuhrau/automationhub/endpoints/api/profile"
	"github.com/gin-gonic/gin"
)

func (s *Service) registerProfileRoutes(api *gin.RouterGroup) {
	profileapi.New(s.accessService()).Register(api.Group("/me"), requireSession)
}
