package api

import (
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/gin-gonic/gin"
)

// registerSSE serves the event stream of the broker.
func registerSSE(api *gin.RouterGroup, broker *sse.Broker) {
	sseApi := api.Group("/sse")
	// with and without trailing slash, so clients don't get a redirect first
	sseApi.GET("", sse.HeadersMiddleware(), broker.ServeHTTP())
	sseApi.GET("/", sse.HeadersMiddleware(), broker.ServeHTTP())
}
