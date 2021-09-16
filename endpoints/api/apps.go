package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
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

func (s *ApiService) createApp(session *Session, c *gin.Context) {
	var app models.App
	c.Bind(&app)
	c.JSON(http.StatusOK, app)
}

func (s *ApiService) uploadApp(session *Session, c *gin.Context) {
	file, err := c.FormFile("test_target")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	filename := filepath.Base(file.Filename)
	filePath := filepath.Join("upload", filename)
	os.MkdirAll("upload", os.ModePerm)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"filename": filename,
		"app_path": filePath,
	})
}