package api

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"net/http"
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