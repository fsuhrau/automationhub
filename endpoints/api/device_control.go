package api

import (
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

func (s *Service) serveMCP(c *gin.Context, project *models.Project) {
	s.mcp.ServeHTTP(c.Writer, c.Request, project, requestOwner(c))
}

// serveMCPUpload receives a build for an upload started with begin_upload.
func (s *Service) serveMCPUpload(c *gin.Context, project *models.Project) {
	s.mcp.ServeUpload(c.Writer, c.Request, project, requestOwner(c), c.Param("upload_id"))
}

// requestOwner names the authenticated caller for device leases.
func requestOwner(c *gin.Context) string {
	value, _ := c.Get(authentication.UserKey)
	switch u := value.(type) {
	case string:
		if u != "" {
			return "token:" + u
		}
	case authentication.User:
		if u.Name != "" {
			return u.Name
		}
		if u.Login != "" {
			return u.Login
		}
	case *authentication.User:
		if u != nil && u.Name != "" {
			return u.Name
		}
		if u != nil && u.Login != "" {
			return u.Login
		}
	}
	return "anonymous"
}
