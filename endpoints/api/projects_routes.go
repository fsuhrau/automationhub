package api

import (
	projectsapi "github.com/fsuhrau/automationhub/endpoints/api/projects"
	"github.com/fsuhrau/automationhub/service/project"
	"github.com/fsuhrau/automationhub/storage/projectstore"
	"github.com/gin-gonic/gin"
)

func (s *Service) registerProjectRoutes(api *gin.RouterGroup) {
	svc := project.New(project.Deps{Repo: projectstore.New(s.db)})
	projectsapi.New(svc).Register(api)
}
