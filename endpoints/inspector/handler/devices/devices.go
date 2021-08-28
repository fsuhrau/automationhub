package devices

import (
	"github.com/fsuhrau/automationhub/endpoints/inspector/handler/visitor"
	"github.com/fsuhrau/automationhub/hub/manager"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(m manager.Devices) func(*gin.Context) {
	return func(c *gin.Context) {
		deviceList := visitor.DeviceList(m)
		c.HTML(http.StatusOK, "devices/index", gin.H{
			"devices": deviceList,
		})
	}
}
