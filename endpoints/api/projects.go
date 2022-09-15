package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

func (s *Service) getProject(identifier string) (*models.Project, error) {
	var project models.Project
	if err := s.db.Where("identifier = ?", identifier).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
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

	request.Identifier = html.EscapeString(strings.Replace(request.Name, " ", "-", -1))

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
