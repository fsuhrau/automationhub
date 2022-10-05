package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

func (s *Service) ResolveProject(context *gin.Context) {
	projectId := context.Param("project_id")
	var project models.Project
	if err := s.db.Where("identifier = ?", projectId).First(&project).Error; err != nil {
		s.error(context, http.StatusNotFound, err)
		return
	}
	context.Set("project", &project)
}

func (s *Service) WithProject(wrapperFunction func(*gin.Context, *models.Project)) func(*gin.Context) {
	return func(context *gin.Context) {
		p, _ := context.Get("project")
		project := p.(*models.Project)
		wrapperFunction(context, project)
	}
}

func (s *Service) getProjects(c *gin.Context) {
	var projects []models.Project
	if err := s.db.Preload("Apps").Find(&projects).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, projects)
}

func (s *Service) createProject(c *gin.Context) {
	var request models.Project

	if err := c.Bind(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	request.Name = strings.TrimSpace(request.Name)

	if len(request.Name) == 0 {
		s.error(c, http.StatusBadRequest, fmt.Errorf("missing name"))
		return
	}

	request.Identifier = html.EscapeString(strings.ToLower(strings.Replace(request.Name, " ", "-", -1)))

	tx := s.db.Begin()
	defer func() {
		if tx.Error != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&request).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, request)
}

func (s *Service) updateProject(c *gin.Context) {
	projectId := c.Param("project_id")

	var request models.Project

	if err := c.Bind(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	var project models.Project
	if err := s.db.First(&project, "Identifier = ?", projectId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	project.Name = request.Name
	if project.Identifier == "default_project" {
		project.Identifier = request.Identifier
	}

	if err := s.db.Save(&project).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, project)
}

func (s *Service) deleteProject(c *gin.Context) {
	projectId := c.Param("project_id")

	var project models.Project
	if err := s.db.First(&project, projectId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	if err := s.db.Delete(&project).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}
