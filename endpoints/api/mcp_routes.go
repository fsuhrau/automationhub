package api

import (
	hubmcp "github.com/fsuhrau/automationhub/endpoints/mcp"
	devicesvc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/devicestore"
	"github.com/gin-gonic/gin"
)

// newMCPHandler builds the MCP tools on the services of this master.
func (s *Service) newMCPHandler() *hubmcp.Handler {
	runs := s.runService()
	return hubmcp.NewHandler(hubmcp.Deps{
		Runs:      runs,
		Protocols: runs,
		Lookup:    runs,
		Apps:      s.appService(),
		Tests:     s.testService(),
		Devices: devicesvc.New(devicesvc.Deps{
			Repo:    devicestore.New(s.db),
			Devices: s.devicesManager,
			Leases:  s.control,
		}),
		Pool:    s.devicesManager,
		Control: s.control,
		Tracker: s.runTracker,
		NodeURL: s.nodeUrl,
	})
}

// registerMCPRoutes adds MCP (streamable HTTP): POST for messages, GET for
// the event stream, DELETE ends a session.
func (s *Service) registerMCPRoutes(projectApi *gin.RouterGroup) {
	projectApi.POST("/mcp", s.WithProject(s.serveMCP))
	projectApi.GET("/mcp", s.WithProject(s.serveMCP))
	projectApi.DELETE("/mcp", s.WithProject(s.serveMCP))
	projectApi.PUT("/mcp/uploads/:upload_id", s.WithProject(s.serveMCPUpload))
}
