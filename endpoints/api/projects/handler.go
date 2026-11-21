// Package projects is the HTTP adapter of the project use cases.
package projects

import (
	"context"
	"net/http"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/service/project"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// Service is what the handlers need from the project use cases.
type Service interface {
	List(ctx context.Context) ([]models.Project, error)
	Create(ctx context.Context, name string) (*models.Project, error)
	Rename(ctx context.Context, identifier string, change project.Change) (*models.Project, error)
	Delete(ctx context.Context, id uint) error
}

// Handler serves the project routes.
type Handler struct {
	svc Service
}

// New returns a Handler on the service.
func New(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds the routes to the api group.
func (h *Handler) Register(api gin.IRoutes) {
	api.GET("/projects", h.list)
	api.POST("/project", h.create)
	api.PUT("/project/:project_id", h.rename)
	api.DELETE("/project/:project_id", h.remove)
}

// bindJSON reads the request body into request; a body that doesn't fit is
// an Invalid error and the response is already written.
func bindJSON(c *gin.Context, request any) bool {
	if err := c.ShouldBindJSON(request); err != nil {
		httpx.Fail(c, apperr.Invalid("%v", err))
		return false
	}
	return true
}

func (h *Handler) list(c *gin.Context) {
	projects, err := h.svc.List(c.Request.Context())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, projects)
}

// create takes only the name from the request: binding the whole project
// let a request create access tokens and apps along with it.
func (h *Handler) create(c *gin.Context) {
	var request struct {
		Name string `json:"name"`
	}
	if !bindJSON(c, &request) {
		return
	}
	created, err := h.svc.Create(c.Request.Context(), request.Name)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) rename(c *gin.Context) {
	var request struct {
		Name       string `json:"name"`
		Identifier string `json:"identifier"`
	}
	if !bindJSON(c, &request) {
		return
	}
	change := project.Change{Name: request.Name, Identifier: request.Identifier}
	renamed, err := h.svc.Rename(c.Request.Context(), c.Param("project_id"), change)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, renamed)
}

func (h *Handler) remove(c *gin.Context) {
	id, err := httpx.ParseID(c, "project_id")
	if err == nil {
		err = h.svc.Delete(c.Request.Context(), id)
	}
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.Status(http.StatusOK)
}
