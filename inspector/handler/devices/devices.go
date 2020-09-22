package devices

import (
	"net/http"

	"github.com/fsuhrau/automationhub/inspector/handler/manager"
	"github.com/gin-gonic/gin"
)

func Index(m manager.DeviceManager) func(*gin.Context) {
	return func(c *gin.Context) {

		deviceList := manager.DeviceList(m)

		c.HTML(http.StatusOK, "devices/index", gin.H{
			"devices": deviceList,
		})
	}
}
