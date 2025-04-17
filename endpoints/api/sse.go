package api

import (
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/gin-gonic/gin"
	"io"
)

func (s *Service) PublishEvent(event sse.Event) {
	s.sseBroker.Events <- event
}

func (s *Service) initSSE(api *gin.RouterGroup) {
	RegisterEventDeviceStatusListener(s)
	RegisterEventAppCreatedListener(s)
	RegisterEventTestCreatedListener(s)
	RegisterNewTestLogEntryListener(s)
	RegisterNewTestProtocolListener(s)
	RegisterNewTestProtocolLogListener(s)
	RegisterEventTestRunFinishedListener(s)

	sseApi := api.Group("/sse")
	sseApi.GET("/", sse.HeadersMiddleware(), s.sseBroker.ServeHTTP(), func(c *gin.Context) {
		cha, ok := c.Get("SSE")
		if !ok {
			return
		}

		clientChannel, ok := cha.(sse.ClientChan)
		if !ok {
			return
		}

		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-clientChannel; ok {
				c.SSEvent(msg.Channel, msg.Content)
				return true
			}
			return false
		})
	})
}
