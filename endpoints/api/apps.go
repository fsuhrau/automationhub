package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) ResolveApp(context *gin.Context) {
	appId := context.Param("app_id")

	p, _ := context.Get("project")
	project := p.(*models.Project)

	var application models.App
	if err := s.db.Preload("Parameter").First(&application, "project_id = ? and id = ?", project.ID, appId).Error; err != nil {
		s.error(context, http.StatusNotFound, err)
		return
	}
	context.Set("app", &application)
}

func (s *Service) WithApp(wrapperFunction func(*gin.Context, *models.Project, *models.App)) func(*gin.Context) {
	return func(context *gin.Context) {
		p, _ := context.Get("project")
		project := p.(*models.Project)
		a, _ := context.Get("app")
		application := a.(*models.App)
		wrapperFunction(context, project, application)
	}
}

func (s *Service) getApps(c *gin.Context, project *models.Project) {

	var apps []models.App
	if err := s.db.Where("project_id = ?", project.ID).Preload("Parameter").Find(&apps).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, apps)
}

func (s *Service) getBinaries(c *gin.Context, project *models.Project, application *models.App) {
	var apps []models.AppBinary
	if err := s.db.Where("app_id = ?", application.ID).Find(&apps).Error; err != nil {
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

func (s *Service) deleteBinary(c *gin.Context, project *models.Project, application *models.App) {
	binaryId := c.Param("binary_id")
	var binary models.AppBinary
	if err := s.db.First(&binary, "app_id = ? and id = ?", application.ID, binaryId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	appPath := filepath.Join(apps.AppStoragePath, binary.AppPath)
	if err := os.RemoveAll(appPath); err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	if err := s.db.Delete(&binary).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) createApp(c *gin.Context, project *models.Project) {
	var application models.App
	if err := c.Bind(&application); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	application.ProjectID = project.ID
	if err := s.db.Create(&application).Error; err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, application)
}

func (s *Service) updateApp(c *gin.Context, project *models.Project, application *models.App) {
	var request models.App
	if err := c.Bind(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	application.Name = request.Name

	// allow change it once since it's a default generated one
	if application.Identifier == "default_app" {
		application.Identifier = request.Identifier
	}

	if err := s.db.Save(&application).Error; err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, application)
}

func (s *Service) addAppParameter(c *gin.Context, project *models.Project, application *models.App) {
	var request models.AppParameter
	if err := c.Bind(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	for _, p := range application.Parameter {
		if p.Name == request.Name {
			s.error(c, http.StatusBadRequest, fmt.Errorf("parameter %s already exists", p.Name))
			return
		}
	}

	request.AppID = application.ID

	if err := s.db.Create(&request).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, request)
}

func (s *Service) removeAppParameter(c *gin.Context, project *models.Project, application *models.App) {

	paramId := c.Param("parameter_id")

	if err := s.db.Delete(&models.AppParameter{}, "app_id = ? and id = ?", application.ID, paramId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (s *Service) updateAppParameter(c *gin.Context, project *models.Project, application *models.App) {
	var request models.AppParameter
	if err := c.Bind(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	paramId := c.Param("parameter_id")
	var param models.AppParameter
	if err := s.db.First(&param, "app_id = ? and id = ?", application.ID, paramId).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	param.Type = request.Type
	param.Name = request.Name

	if err := s.db.Save(&param).Error; err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, application)
}

func (s *Service) updateBinary(c *gin.Context, project *models.Project, application *models.App) {
	binaryId := c.Param("binary_id")

	var newApp models.AppBinary
	c.Bind(&newApp)

	var binary models.AppBinary
	if err := s.db.Find(&binary, "app_id = ? and id = ?", application.ID, binaryId).Error; err != nil {
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

func (s *Service) uploadBinary(c *gin.Context, project *models.Project, application *models.App) {
	file, err := c.FormFile("test_target")

	if err != nil {
		s.error(c, http.StatusBadRequest, fmt.Errorf("get form err: %s", err.Error()))
		return
	}

	filename := filepath.Base(file.Filename)
	filePath := filepath.Join(apps.AppStoragePath, filename)

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

	ext := filepath.Ext(filename)
	newFileName := strings.TrimSuffix(filename, ext)
	params.App.AppPath = fmt.Sprintf("%s_%s%s", newFileName, params.App.Hash, ext)
	newFilePath := filepath.Join(apps.AppStoragePath, params.App.AppPath)
	if err := os.Rename(filePath, newFilePath); err != nil {
		s.error(c, http.StatusInternalServerError, fmt.Errorf("unable to store file err: %s", err.Error()))
		return
	}

	a := models.AppBinary{}
	result := s.db.First(&a, "hash = '%s'", params.App.Hash)
	if result.RowsAffected > 0 {
		s.error(c, http.StatusBadRequest, fmt.Errorf("app exists already "))
		return
	}
	if err != nil && err == gorm.ErrRecordNotFound {
		s.error(c, http.StatusBadRequest, fmt.Errorf("app exists already err: %v", err))
		return
	}

	var launchActivity string
	if params.App.Android != nil {
		launchActivity = params.App.Android.LaunchActivity
	}
	var executable string
	if params.App.Executable != nil {
		executable = params.App.Executable.Executable
	}

	appDto := models.AppBinary{
		AppPath:    params.App.AppPath,
		Hash:       params.App.Hash,
		Name:       params.Name,
		Identifier: params.Identifier,
		AppID:      application.ID,
		Version:    params.Version,
		Android: models.Android{
			LaunchActivity: launchActivity,
		},
		Executable: models.Executable{
			Executable: executable,
		},
		Platform:   params.Platform,
		Additional: params.App.Additional,
		Size:       params.App.Size,
	}
	if err := s.db.Create(&appDto).Error; err != nil {
		s.error(c, http.StatusInternalServerError, fmt.Errorf("unable to create app err: %s", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, appDto)
}
