package api

import (
	protocolsapi "github.com/fsuhrau/automationhub/endpoints/api/protocols"
	devicesvc "github.com/fsuhrau/automationhub/service/device"
	protocolsvc "github.com/fsuhrau/automationhub/service/protocol"
	"github.com/fsuhrau/automationhub/storage/runstore"
	"github.com/gin-gonic/gin"
)

// protocolService builds the protocol use cases on the database; a run is
// checked by the run service.
func (s *Service) protocolService() *protocolsvc.Service {
	return protocolsvc.New(protocolsvc.Deps{
		Repo:       runstore.New(s.db),
		Scope:      s.runService(),
		DeviceName: devicesvc.DisplayName,
	})
}

func (s *Service) registerProtocolRoutes(appApi *gin.RouterGroup) {
	protocolsapi.New(s.protocolService()).RegisterProtocols(appApi)
}
