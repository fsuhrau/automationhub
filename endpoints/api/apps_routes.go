package api

import (
	appsapi "github.com/fsuhrau/automationhub/endpoints/api/apps"
	"github.com/gin-gonic/gin"
)

func (s *Service) registerAppRoutes(projectApi *gin.RouterGroup) {
	appsapi.New(s.appService()).RegisterProject(projectApi)
}

// registerAppDetailRoutes adds the routes of one app: the app itself and
// its parameters.
func (s *Service) registerAppDetailRoutes(appApi *gin.RouterGroup) {
	appsapi.New(s.appService()).RegisterApp(appApi)
}
