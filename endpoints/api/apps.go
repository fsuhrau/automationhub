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

func (s *ApiService) getApps(session *Session, c *gin.Context) {
	var apps []models.App
	if err := s.db.Find(&apps).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, apps)
}

func (s *ApiService) getApp(session *Session, c *gin.Context) {
	appId := c.Param("app_id")
	var app models.App
	if err := s.db.First(&app, appId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, app)
}

func (s *ApiService) deleteApp(session *Session, c *gin.Context) {
	appId := c.Param("app_id")
	var app models.App
	if err := s.db.First(&app, appId).Error; err != nil {
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

func (s *ApiService) getAppFunctions(session *Session, c *gin.Context) {
	appId := c.Param("app_id")
	var funcs []models.AppFunction
	if err := s.db.Find(&funcs, "app_id = ?", appId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, funcs)
}

func (s *ApiService) createApp(session *Session, c *gin.Context) {
	var app models.App
	c.Bind(&app)
	c.JSON(http.StatusOK, app)
}

func (s *ApiService) updateApp(session *Session, c *gin.Context) {
	appId := c.Param("app_id")

	var newApp models.App
	c.Bind(&newApp)

	var app models.App

	if err := s.db.Find(&app, appId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	app.Tags = newApp.Tags
	if err := s.db.Save(&app).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, app)
}

func (s *ApiService) uploadApp(session *Session, c *gin.Context) {
	file, err := c.FormFile("test_target")
	if err != nil {
		s.error(c, http.StatusBadRequest, fmt.Errorf("get form err: %s", err.Error()))
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

	a := models.App{}
	result := s.db.First(&a, "hash = '%s'", params.Hash)
	if result.RowsAffected > 0 {
		s.error(c, http.StatusBadRequest, fmt.Errorf("app exists already "))
		return
	}
	if err != nil && err == gorm.ErrRecordNotFound {
		s.error(c, http.StatusBadRequest, fmt.Errorf("app exists already err: %v", err))
		return
	}

	appDto := models.App{
		AppPath: params.AppPath,
		Hash: params.Hash,
		Name: params.Name,
		Identifier: params.Identifier,
		AppID: params.Identifier,
		Version: params.Version,
		LaunchActivity: params.LaunchActivity,
		Platform: params.Platform,
		Additional: params.Additional,
		Size: params.Size,
	}
	if err := s.db.Create(&appDto).Error; err != nil {
		s.error(c, http.StatusInternalServerError, fmt.Errorf("unable to create app err: %s", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, appDto)
}