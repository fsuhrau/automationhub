// Package apps is the HTTP adapter of the apps and their parameters: it
// reads the request, calls the app service and writes the result.
package apps

import (
	"context"
	"net/http"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// Service is what the handlers call.
type Service interface {
	Apps(ctx context.Context, projectID uint) ([]models.App, error)
	CreateApp(ctx context.Context, projectID uint, in appsvc.AppInput) (*models.App, error)
	UpdateApp(ctx context.Context, sc appsvc.Scope, in appsvc.AppInput) (*models.App, error)
	AddParameter(ctx context.Context, sc appsvc.Scope, in appsvc.ParameterInput) (*models.AppParameter, error)
	UpdateParameter(ctx context.Context, ref appsvc.ParameterRef, in appsvc.ParameterInput) (*models.AppParameter, error)
	RemoveParameter(ctx context.Context, ref appsvc.ParameterRef) error
}

// Handler serves the app routes.
type Handler struct {
	svc Service
}

// New builds the handler on the service.
func New(svc Service) *Handler { return &Handler{svc: svc} }

// RegisterProject adds the routes below a project.
func (h *Handler) RegisterProject(r *gin.RouterGroup) {
	r.POST("/app", h.create)
	r.GET("/apps", h.list)
}

// RegisterApp adds the routes of one app, below the group that resolves it.
func (h *Handler) RegisterApp(r *gin.RouterGroup) {
	r.PUT("/", h.update)
	r.POST("/parameter", h.addParameter)
	r.PUT("/parameter/:parameter_id", h.updateParameter)
	r.DELETE("/parameter/:parameter_id", h.removeParameter)
}

func scope(c *gin.Context) appsvc.Scope {
	return appsvc.Scope{ProjectID: httpx.Project(c).ID, AppID: httpx.App(c).ID}
}

func (h *Handler) list(c *gin.Context) {
	found, err := h.svc.Apps(c.Request.Context(), httpx.Project(c).ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, found)
}

func (h *Handler) create(c *gin.Context) {
	var in appsvc.AppInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Fail(c, invalid(err))
		return
	}
	created, err := h.svc.CreateApp(c.Request.Context(), httpx.Project(c).ID, in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, created)
}

func (h *Handler) update(c *gin.Context) {
	var in appsvc.AppInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Fail(c, invalid(err))
		return
	}
	updated, err := h.svc.UpdateApp(c.Request.Context(), scope(c), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}
