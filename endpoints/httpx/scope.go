package httpx

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// The context keys under which the ResolveProject and ResolveApp middlewares
// store what they loaded.
const (
	ProjectKey = "project"
	AppKey     = "app"
)

// Project is the project resolved for the request, nil if there is none.
func Project(c *gin.Context) *models.Project {
	p, _ := c.Get(ProjectKey)
	project, _ := p.(*models.Project)
	return project
}

// App is the app resolved for the request, nil if there is none.
func App(c *gin.Context) *models.App {
	a, _ := c.Get(AppKey)
	application, _ := a.(*models.App)
	return application
}
