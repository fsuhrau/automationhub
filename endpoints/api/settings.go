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
	"strings"
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
	var req struct {
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		s.error(c, http.StatusBadRequest, fmt.Errorf("the token needs a name"))
		return
	}

	id, err := uuid.NewV4()
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	projectId := c.Param("project_id")
	var project models.Project
	if err := s.db.First(&project, "identifier = ?", projectId).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		s.error(c, http.StatusNotFound, fmt.Errorf("project %s not found", projectId))
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
		s.error(c, http.StatusNotFound, fmt.Errorf("project %s not found", projectId))
		return
	}

	var accessTokens []models.AccessToken
	if err := s.db.Find(&accessTokens, "project_id = ?", project.ID).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, accessTokens)
}

// deleteAccessToken revokes a token of the project. It used to delete any
// token by id: other projects', nodes' and personal ones.
func (s *Service) deleteAccessToken(c *gin.Context) {
	var project models.Project
	if err := s.db.First(&project, "identifier = ?", c.Param("project_id")).Error; err != nil {
		s.error(c, http.StatusNotFound, fmt.Errorf("project %s not found", c.Param("project_id")))
		return
	}

	var accessToken models.AccessToken
	if err := s.db.First(&accessToken, "id = ? AND project_id = ?", c.Param("token_id"), project.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.error(c, http.StatusNotFound, fmt.Errorf("token %s not found", c.Param("token_id")))
		} else {
			s.error(c, http.StatusInternalServerError, err)
		}
		return
	}

	if err := s.db.Delete(&accessToken).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// newNode is a created node with its token; the token is only sent here
// and when a new one is issued
type newNode struct {
	models.Node
	Token string `json:"token"`
}

// issueNodeToken replaces the tokens of a node with a new one
func issueNodeToken(tx *gorm.DB, node *models.Node) (string, error) {
	token, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	if err := tx.Where("node_id = ?", node.ID).Delete(&models.AccessToken{}).Error; err != nil {
		return "", err
	}
	if err := tx.Create(&models.AccessToken{NodeID: &node.ID, Token: token.String(), Name: node.Name}).Error; err != nil {
		return "", err
	}
	return token.String(), nil
}

func (s *Service) createNode(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		s.error(c, http.StatusBadRequest, fmt.Errorf("the node needs a name"))
		return
	}

	id, err := uuid.NewV4()
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	created := newNode{Node: models.Node{Identifier: id.String(), Name: req.Name}}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&created.Node).Error; err != nil {
			return err
		}
		token, err := issueNodeToken(tx, &created.Node)
		created.Token = token
		return err
	})
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

// issueNodeTokenHandler gives a node a new token, for setting it up again;
// the old token stops working
func (s *Service) issueNodeTokenHandler(c *gin.Context) {
	var node models.Node
	if err := s.db.First(&node, "id = ?", c.Param("node_id")).Error; err != nil {
		s.error(c, http.StatusNotFound, fmt.Errorf("node %s not found", c.Param("node_id")))
		return
	}
	var token string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var err error
		token, err = issueNodeToken(tx, &node)
		return err
	})
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, newNode{Node: node, Token: token})
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

// deleteNode removes a node and its tokens, so it can't connect anymore
func (s *Service) deleteNode(c *gin.Context) {
	var node models.Node
	if err := s.db.First(&node, "id = ?", c.Param("node_id")).Error; err != nil {
		s.error(c, http.StatusNotFound, fmt.Errorf("node %s not found", c.Param("node_id")))
		return
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("node_id = ?", node.ID).Delete(&models.AccessToken{}).Error; err != nil {
			return err
		}
		return tx.Delete(&node).Error
	})
	if err != nil {
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
