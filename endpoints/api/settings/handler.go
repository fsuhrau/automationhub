// Package settings is the HTTP adapter of the project settings: the users,
// the access tokens of the project and the nodes.
package settings

import (
	"context"
	"net/http"
	"time"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/service/access"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// Service is what the handlers need from the access use cases.
type Service interface {
	Users(ctx context.Context) ([]models.User, error)
	CreateProjectToken(ctx context.Context, projectID uint, in access.NewToken) (*models.AccessToken, error)
	ProjectTokens(ctx context.Context, projectID uint) ([]models.AccessToken, error)
	DeleteProjectToken(ctx context.Context, projectID, tokenID uint) error
	CreateNode(ctx context.Context, name string) (*access.NodeToken, error)
	Nodes(ctx context.Context) ([]models.Node, error)
	Node(ctx context.Context, id uint) (*models.Node, error)
	IssueNodeToken(ctx context.Context, id uint) (*access.NodeToken, error)
	DeleteNode(ctx context.Context, id uint) error
}

// Handler serves the settings routes.
type Handler struct {
	svc Service
}

// New returns a Handler on the service.
func New(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds the routes to the settings group of a project. The group
// has to resolve the project and decide who may use the settings.
func (h *Handler) Register(settings gin.IRoutes) {
	settings.GET("/users", h.users)

	settings.POST("/access_token", h.createToken)
	settings.GET("/access_tokens", h.tokens)
	settings.DELETE("/access_token/:token_id", h.deleteToken)

	settings.POST("/node", h.createNode)
	settings.GET("/nodes", h.nodes)
	settings.GET("/nodes/:node_id", h.node)
	settings.DELETE("/nodes/:node_id", h.deleteNode)
	settings.POST("/nodes/:node_id/token", h.issueNodeToken)
}

// reply is the status a use case answers with when it succeeds.
type reply int

const (
	ok      reply = http.StatusOK
	created reply = http.StatusCreated
)

// with writes the value of a use case, or its error.
func (r reply) with(c *gin.Context, value any, err error) {
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(int(r), value)
}

// done answers a delete: 200 without a body, or the error.
func done(c *gin.Context, err error) {
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) users(c *gin.Context) {
	users, err := h.svc.Users(c.Request.Context())
	ok.with(c, users, err)
}

func (h *Handler) createToken(c *gin.Context) {
	var request struct {
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	if err := httpx.BindJSON(c, &request); err != nil {
		httpx.Fail(c, err)
		return
	}
	token, err := h.svc.CreateProjectToken(c.Request.Context(), httpx.Project(c).ID, access.NewToken(request))
	created.with(c, token, err)
}

func (h *Handler) tokens(c *gin.Context) {
	tokens, err := h.svc.ProjectTokens(c.Request.Context(), httpx.Project(c).ID)
	ok.with(c, tokens, err)
}

func (h *Handler) deleteToken(c *gin.Context) {
	id, err := httpx.ParseID(c, "token_id")
	if err == nil {
		err = h.svc.DeleteProjectToken(c.Request.Context(), httpx.Project(c).ID, id)
	}
	done(c, err)
}
