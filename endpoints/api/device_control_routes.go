package api

import (
	controlapi "github.com/fsuhrau/automationhub/endpoints/control"
	"github.com/gin-gonic/gin"
)

// deviceControl is the live control handler, which resolves devices and
// apps through the device handler.
func (e *Endpoint) deviceControl() *controlapi.Handler {
	d := e.deps
	return controlapi.NewHandler(d.Control, e.devices.ResolveControlDevice, e.devices.ResolveControlApp, requestOwner, d.NodeURL).
		WithAllowedOrigins(d.Cors)
}

func registerDeviceControlRoutes(projectApi *gin.RouterGroup, handler *controlapi.Handler) {
	handler.Register(projectApi)
}
