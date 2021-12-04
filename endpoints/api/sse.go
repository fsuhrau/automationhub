package api

import (
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/gin-gonic/gin"
	"io"
	"log"
)

func (s *ApiService) PublishEvent(event sse.Event) {
	s.sseBroker.Events <- event
}

func (s *ApiService) initSSE(api *gin.RouterGroup) {
	RegisterEventDeviceStatusListener(s)
	RegisterEventAppCreatedListener(s)
	RegisterEventTestCreatedListener(s)
	RegisterNewTestLogEntryListener(s)
	RegisterNewTestProtocolListener(s)
	RegisterNewTestProtocolLogListener(s)
	RegisterEventTestRunFinishedListener(s)

	sseApi := api.Group("/sse")
	sseApi.Use(sse.HeadersMiddleware())
	sseApi.Use(s.sseBroker.ServeHTTP())
	sseApi.GET("/", func(c *gin.Context) {
		log.Printf("register for streaming")
		cha, _ := c.Get("SSE")
		clientChannel, _ := cha.(sse.ClientChan)
		c.Stream(func(w io.Writer) bool {
			msg := <-clientChannel
			c.SSEvent(msg.Channel, msg.Content)
			return true
		})
	})
}
