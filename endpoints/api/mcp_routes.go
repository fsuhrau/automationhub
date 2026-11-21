package api

import (
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	hubmcp "github.com/fsuhrau/automationhub/endpoints/mcp"
	"github.com/gin-gonic/gin"
)

// registerMCPRoutes adds MCP (streamable HTTP): POST for messages, GET for
// the event stream, DELETE ends a session.
func registerMCPRoutes(projectApi *gin.RouterGroup, mcp *hubmcp.Handler) {
	serve := func(c *gin.Context) {
		mcp.ServeHTTP(c.Writer, c.Request, httpx.Project(c), requestOwner(c))
	}
	upload := func(c *gin.Context) {
		mcp.ServeUpload(c.Writer, c.Request, httpx.Project(c), requestOwner(c), c.Param("upload_id"))
	}
	projectApi.POST("/mcp", serve)
	projectApi.GET("/mcp", serve)
	projectApi.DELETE("/mcp", serve)
	projectApi.PUT("/mcp/uploads/:upload_id", upload)
}
