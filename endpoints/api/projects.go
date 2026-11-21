package api

import (
	"net/http"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

func (s *Service) ResolveProject(context *gin.Context) {
	projectId := context.Param("project_id")
	var project models.Project
	if err := s.db.Where("identifier = ?", projectId).First(&project).Error; err != nil {
		s.error(context, http.StatusNotFound, err)
		return
	}
	context.Set(httpx.ProjectKey, &project)
}

func (s *Service) WithProject(wrapperFunction func(*gin.Context, *models.Project)) func(*gin.Context) {
	return func(context *gin.Context) {
		p, _ := context.Get(httpx.ProjectKey)
		project := p.(*models.Project)
		wrapperFunction(context, project)
	}
}
