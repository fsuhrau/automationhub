package devices

import (
	"context"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/gin-gonic/gin"
)

// ResolveControlDevice returns the live device for :device_id within the
// project of the request, for the live control routes. It writes the error
// response itself when there is none.
func (h *Handler) ResolveControlDevice(c *gin.Context) (device.Device, bool) {
	project := httpx.Project(c)
	if project == nil {
		httpx.Fail(c, apperr.NotFound("project not found"))
		return nil, false
	}
	dev, err := h.svc.ControlDevice(detached(c), project, c.Param("device_id"))
	if err != nil {
		httpx.Fail(c, err)
		return nil, false
	}
	return dev, true
}

// ResolveControlApp loads an app bundle that belongs to the project of the
// request.
func (h *Handler) ResolveControlApp(c *gin.Context, binaryID uint) (*app.Parameter, error) {
	project := httpx.Project(c)
	if project == nil {
		return nil, apperr.NotFound("project not found")
	}
	return h.svc.ControlApp(detached(c), project, binaryID)
}

// detached is the request context without its cancellation. The control
// routes resolve the device first and then stream until the client leaves;
// a client that is gone already must still get its stream, which ends at
// once, not an error from a lookup that was cancelled.
func detached(c *gin.Context) context.Context {
	return context.WithoutCancel(c.Request.Context())
}
