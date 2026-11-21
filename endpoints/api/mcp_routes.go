package api

import "github.com/gin-gonic/gin"

// registerMCPRoutes adds MCP (streamable HTTP): POST for messages, GET for
// the event stream, DELETE ends a session.
func (s *Service) registerMCPRoutes(projectApi *gin.RouterGroup) {
	projectApi.POST("/mcp", s.WithProject(s.serveMCP))
	projectApi.GET("/mcp", s.WithProject(s.serveMCP))
	projectApi.DELETE("/mcp", s.WithProject(s.serveMCP))
	projectApi.PUT("/mcp/uploads/:upload_id", s.WithProject(s.serveMCPUpload))
}
