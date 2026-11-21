package api

import (
	protocolsapi "github.com/fsuhrau/automationhub/endpoints/api/protocols"
	"github.com/gin-gonic/gin"
)

func (s *Service) registerBaselineRoutes(appApi *gin.RouterGroup) {
	protocolsapi.New(s.protocolService()).RegisterBaselines(appApi)
}
