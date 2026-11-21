package api

import (
	devicesapi "github.com/fsuhrau/automationhub/endpoints/api/devices"
	devicesvc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/devicestore"
	"github.com/fsuhrau/automationhub/tester/unity"
	"github.com/gin-gonic/gin"
)

// deviceHandler builds the device service on this service's database, pool
// and leases. extractParams and defaultNaN stay here until the endpoints
// that own them move to services.
func (s *Service) deviceHandler() *devicesapi.Handler {
	return devicesapi.New(devicesvc.New(devicesvc.Deps{
		Repo:    devicestore.New(s.db),
		Devices: s.devicesManager,
		Leases:  s.control,
		NewExecutor: func() devicesvc.Executor {
			return unity.NewExecutor(s.devicesManager, nil)
		},
		ExtractParams: extractParams,
		CleanProtocol: defaultNaN,
	}))
}

func (s *Service) registerDeviceRoutes(projectApi *gin.RouterGroup) {
	s.deviceHandler().Register(projectApi)
}
