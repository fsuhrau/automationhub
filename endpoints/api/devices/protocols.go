package devices

import (
	"net/http"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/gin-gonic/gin"
)

const (
	protocolsDefault = 10
	protocolsMax     = 50
)

// RegisterProtocols adds the route that lists the latest test results on a
// device, for the device page.
func (h *Handler) RegisterProtocols(project *gin.RouterGroup) {
	project.GET("/device/:device_id/protocols", h.protocols)
}

func (h *Handler) protocols(c *gin.Context) {
	limit := httpx.Limit(c, protocolsDefault, protocolsMax)
	query := svc.ProtocolQuery{DeviceID: c.Param("device_id"), Limit: limit}
	protocols, err := h.svc.Protocols(c.Request.Context(), httpx.Project(c), query)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, protocols)
}
