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
	RegisterEventAppCreatedListener(s)
	RegisterEventTestCreatedListener(s)
	RegisterNewTestProtocolListener(s)
	RegisterNewTestProtocolLogListener(s)
	RegisterNewTestLogEntryListener(s)
	RegisterEventTestRunFinishedListener(s)

	sseApi := api.Group("/sse")
	sseApi.GET("/", sse.HeadersMiddleware(), s.sseBroker.ServeHTTP())
}
