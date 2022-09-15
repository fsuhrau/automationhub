package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) getApp(appId string) (*models.App, error) {
	var app models.App
	if err := s.db.First(&app, appId).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (s *Service) getApps(c *gin.Context) {
	projectId := c.Param("project_id")
	project, err := s.getProject(projectId)
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	var apps []models.App
	if err := s.db.Where("project_id = ?", project.ID).Find(&apps).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, apps)
}

func (s *Service) getBinaries(c *gin.Context) {
	appId := c.Param("app_id")
	var apps []models.AppBinary
	if err := s.db.Where("app_id = ?", appId).Find(&apps).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, apps)
}

/*
func (s *Service) getApp(c *gin.Context) {
	appId := c.Param("app_id")
	var app models.AppBinary
	if err := s.db.First(&app, appId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, app)
}
*/

func (s *Service) deleteBinary(c *gin.Context) {
	binaryId := c.Param("binary_id")
	var app models.AppBinary
	if err := s.db.First(&app, binaryId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	if err := os.RemoveAll(app.AppPath); err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	if err := s.db.Delete(&app).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) createApp(c *gin.Context) {
	projectId := c.Param("project_id")
	project, err := s.getProject(projectId)
	if err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	var app models.App
	c.Bind(&app)
	app.ProjectID = project.ID
	if err := s.db.Create(&app).Error; err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, app)
}

func (s *Service) updateBinary(c *gin.Context) {
	binaryId := c.Param("binary_id")

	var newApp models.AppBinary
	c.Bind(&newApp)

	var binary models.AppBinary

	if err := s.db.Find(&binary, binaryId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	binary.Tags = newApp.Tags
	if err := s.db.Save(&binary).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, binary)
}

func (s *Service) uploadBinary(c *gin.Context) {
	file, err := c.FormFile("test_target")
	projectId := c.Param("project_id")

	if err != nil {
		s.error(c, http.StatusBadRequest, fmt.Errorf("get form err: %s", err.Error()))
		return
	}

	project, err := s.getProject(projectId)
	if err != nil {
		s.error(c, http.StatusInternalServerError, fmt.Errorf("project not found err: %s", err.Error()))
		return
	}

	filename := filepath.Base(file.Filename)
	filePath := filepath.Join("upload", filename)

	os.MkdirAll("upload", os.ModePerm)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		s.error(c, http.StatusInternalServerError, fmt.Errorf("upload file err: %s", err.Error()))
		return
	}

	analyser := app.NewAnalyser(filePath)
	if err := analyser.AnalyseFile(); err != nil {
		s.error(c, http.StatusInternalServerError, fmt.Errorf("unable to analyse app err: %s", err.Error()))
		return
	}
	params := analyser.GetParameter()

	ext := filepath.Ext(filePath)
	newPath := strings.TrimSuffix(filePath, ext)
	params.AppPath = fmt.Sprintf("%s_%s%s", newPath, params.Hash, ext)
	if err := os.Rename(filePath, params.AppPath); err != nil {
		s.error(c, http.StatusInternalServerError, fmt.Errorf("unable to store file err: %s", err.Error()))
		return
	}

	var app models.App
	if err := s.db.Where("project_id = ? and identifier = ?", project.ID, params.Identifier).First(&app).Error; err != nil {
		s.error(c, http.StatusBadRequest, fmt.Errorf("no app defined for identifer %s", params.Identifier))
		return
	}

	a := models.AppBinary{}
	result := s.db.First(&a, "hash = '%s'", params.Hash)
	if result.RowsAffected > 0 {
		s.error(c, http.StatusBadRequest, fmt.Errorf("app exists already "))
		return
	}
	if err != nil && err == gorm.ErrRecordNotFound {
		s.error(c, http.StatusBadRequest, fmt.Errorf("app exists already err: %v", err))
		return
	}

	appDto := models.AppBinary{
		AppPath:        params.AppPath,
		Hash:           params.Hash,
		Name:           params.Name,
		Identifier:     params.Identifier,
		AppID:          app.ID,
		Version:        params.Version,
		LaunchActivity: params.LaunchActivity,
		Platform:       params.Platform,
		Additional:     params.Additional,
		Size:           params.Size,
	}
	if err := s.db.Create(&appDto).Error; err != nil {
		s.error(c, http.StatusInternalServerError, fmt.Errorf("unable to create app err: %s", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, appDto)
}
