package master

import (
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *nodeMaster) Action(c *gin.Context) {
	type request struct {
		Node     string           `json:"node"`
		DeviceID string           `json:"device_id"`
		Response *action.Response `json:"response"`
	}

	var req request

	if err := c.BindJSON(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return
	}

	if handler := device.ActionHandlers(); handler != nil {
		for i := range handler {
			handler[i].OnActionResponse(device, req.Response)
		}
	}
}
