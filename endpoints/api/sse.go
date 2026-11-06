package api

import (
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/gin-gonic/gin"
)

func (s *Service) PublishEvent(event sse.Event) {
	s.sseBroker.Publish(event)
}

func (s *Service) initSSE(api *gin.RouterGroup) {
	RegisterEventDeviceStatusListener(s)
	RegisterEventDeviceListListener(s)
	RegisterEventDeviceLockListener(s)
	RegisterEventAppCreatedListener(s)
	RegisterEventTestCreatedListener(s)
	RegisterNewTestProtocolListener(s)
	RegisterNewTestProtocolLogListener(s)
	RegisterNewTestLogEntryListener(s)
	RegisterEventTestRunFinishedListener(s)
	RegisterNodeStatusListener(s)

	sseApi := api.Group("/sse")
	// with and without trailing slash, so clients don't get a redirect first
	sseApi.GET("", sse.HeadersMiddleware(), s.sseBroker.ServeHTTP())
	sseApi.GET("/", sse.HeadersMiddleware(), s.sseBroker.ServeHTTP())
}
