package master

import (
	"net/http"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-gonic/gin"
)

// DevicesChanged is called by a node when its device list changed, so the
// master refreshes that node's devices right away instead of on the next poll.
func (s *nodeMaster) DevicesChanged(c *gin.Context) {
	var req struct {
		Node string `json:"node" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	events.NodeDevicesChanged.Trigger(events.NodeDevicesChangedPayload{NodeIdentifier: manager.NodeIdentifier(req.Node)})
	c.Status(http.StatusAccepted)
}
