package api

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/gin-gonic/gin"
)

func (s *Service) getData(c *gin.Context) {
	name := c.Param("name")
	if !filepath.IsLocal(name) {
		s.error(c, http.StatusBadRequest, fmt.Errorf("invalid data name %q", name))
		return
	}
	c.File(filepath.Join(apps.TestDataPath, name))
}
