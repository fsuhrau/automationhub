package api

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"net/http"
	"time"
)

func (s *Service) getAccessTokens(c *gin.Context) {
	var accessTokens []models.AccessToken
	if err := s.db.Find(&accessTokens).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, accessTokens)
}

func (s *Service) deleteAccessToken(c *gin.Context) {
	tokenId := c.Param("token_id")

	var accessToken models.AccessToken
	if err := s.db.First(&accessToken, tokenId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	if err := s.db.Delete(&accessToken).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) createAccessToken(c *gin.Context) {
	type Request struct {
		Name      string
		ExpiresAt *time.Time
	}
	var req Request

	c.Bind(&req)

	id, err := uuid.NewV4()
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	newToken := models.AccessToken{
		Token:     id.String(),
		ExpiresAt: req.ExpiresAt,
		Name:      req.Name,
	}

	if err := s.db.Create(&newToken).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, newToken)
}
