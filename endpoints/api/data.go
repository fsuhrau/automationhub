package api

import (
	"path/filepath"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/gin-gonic/gin"
)

// registerDataRoutes serves the files of the test data folder.
func registerDataRoutes(api *gin.RouterGroup, dir string) {
	api.GET("/data/:name", func(c *gin.Context) {
		name := c.Param("name")
		if !filepath.IsLocal(name) {
			httpx.Fail(c, apperr.Invalid("invalid data name %q", name))
			return
		}
		c.File(filepath.Join(dir, name))
	})
}
