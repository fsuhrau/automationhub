package master

import (
	"net/http"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/gin-gonic/gin"
)

// DevicesChanged is called by a node when its device list changed. The body
// carries the complete list, which the master applies right away; a body
// without devices (older nodes) makes the master fetch the list over RPC.
func (s *nodeMaster) DevicesChanged(c *gin.Context) {
	var req node.DevicesChangedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payload := events.NodeDevicesChangedPayload{
		NodeIdentifier: manager.NodeIdentifier(req.Node),
		SentAt:         req.SentAt,
	}
	if req.Devices != nil {
		payload.Devices = node.FromDevicesResponse(*req.Devices)
	}
	events.NodeDevicesChanged.Trigger(payload)
	c.Status(http.StatusAccepted)
}
