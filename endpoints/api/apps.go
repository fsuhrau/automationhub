package api

import (
	"net/http"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

func (s *Service) ResolveApp(context *gin.Context) {
	appId := context.Param("app_id")

	p, _ := context.Get(httpx.ProjectKey)
	project := p.(*models.Project)

	var application models.App
	if err := s.db.Preload("Parameter").First(&application, "project_id = ? and id = ?", project.ID, appId).Error; err != nil {
		s.error(context, http.StatusNotFound, err)
		return
	}
	context.Set(httpx.AppKey, &application)
}

func (s *Service) WithApp(wrapperFunction func(*gin.Context, *models.Project, *models.App)) func(*gin.Context) {
	return func(context *gin.Context) {
		p, _ := context.Get(httpx.ProjectKey)
		project := p.(*models.Project)
		a, _ := context.Get(httpx.AppKey)
		application := a.(*models.App)
		wrapperFunction(context, project, application)
	}
}
