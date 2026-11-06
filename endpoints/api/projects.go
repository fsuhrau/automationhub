package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	if err := s.db.Preload("Apps").Preload("Apps.Parameter").Find(&projects).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, projects)
}

// projectSlug turns a name into an identifier that is safe in URLs: lower
// case letters, digits and single dashes
func projectSlug(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case b.Len() > 0 && !dash:
			b.WriteByte('-')
			dash = true
		}
	}
	slug := strings.TrimSuffix(b.String(), "-")
	if len(slug) > 48 {
		slug = strings.TrimSuffix(slug[:48], "-")
	}
	if slug == "" {
		slug = "project"
	}
	return slug
}

// createProject creates a project from its name. Only the name is taken
// from the request: binding the whole project let a request create access
// tokens and apps along with it. The identifier is derived from the name
// and made unique; deleted projects keep theirs, so old links don't lead
// to a new project.
func (s *Service) createProject(c *gin.Context) {
	var request struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		s.error(c, http.StatusBadRequest, fmt.Errorf("the project needs a name"))
		return
	}

	project := models.Project{Name: name}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		base := projectSlug(name)
		project.Identifier = base
		for i := 2; ; i++ {
			var taken int64
			if err := tx.Unscoped().Model(&models.Project{}).Where("identifier = ?", project.Identifier).Count(&taken).Error; err != nil {
				return err
			}
			if taken == 0 {
				break
			}
			project.Identifier = fmt.Sprintf("%s-%d", base, i)
		}
		return tx.Create(&project).Error
	})
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, project)
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
