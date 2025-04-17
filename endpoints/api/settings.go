package api

import (
	"errors"
	"fmt"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func (s *Service) getUsers(c *gin.Context) {
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, users)
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

	projectId := c.Param("project_id")
	var project models.Project
	if err := s.db.First(&project, "identifier = ?", projectId).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		s.error(c, http.StatusNotFound, fmt.Errorf("project "+projectId+" not found"))
		return
	}

	newToken := models.AccessToken{
		ProjectID: &project.ID,
		NodeID:    nil,
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

func (s *Service) getAccessTokens(c *gin.Context) {

	projectId := c.Param("project_id")
	var project models.Project
	if err := s.db.First(&project, "identifier = ?", projectId).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		s.error(c, http.StatusNotFound, fmt.Errorf("project "+projectId+" not found"))
		return
	}

	var accessTokens []models.AccessToken
	if err := s.db.Find(&accessTokens, "project_id = ?", project.ID).Error; err != nil {
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

func (s *Service) createNode(c *gin.Context) {
	type Request struct {
		Name string
	}
	var req Request

	c.Bind(&req)

	id, err := uuid.NewV4()
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	newNode := models.Node{
		Identifier: id.String(),
		Name:       req.Name,
	}

	if err := s.db.Create(&newNode).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	token, err := uuid.NewV4()
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	newToken := models.AccessToken{
		NodeID:    &newNode.ID,
		ProjectID: nil,
		Token:     token.String(),
		ExpiresAt: nil,
		Name:      req.Name,
	}

	if err := s.db.Create(&newToken).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, newNode)
}

func (s *Service) getNodes(c *gin.Context) {
	var nodes []models.Node
	if err := s.db.Find(&nodes).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	for i := range nodes {
		nodes[i].Status, _ = s.nodeManager.GetStatus(manager.NodeIdentifier(nodes[i].Identifier))
	}

	c.JSON(http.StatusOK, nodes)
}

func (s *Service) deleteNode(c *gin.Context) {
	nodeId := c.Param("node_id")

	var node models.Node
	if err := s.db.First(&node, nodeId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	if err := s.db.Delete(&node).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) getNodeStatus(c *gin.Context) {
	nodeId := c.Param("node_id")

	var node models.Node
	if err := s.db.First(&node, nodeId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	// TODO add node status

	c.JSON(http.StatusOK, node)
}
