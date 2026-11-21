// Package tests holds the HTTP handlers of the test definitions of an app:
// they parse the request, call the test service and write the answer.
package tests

import (
	"context"
	"net/http"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	service "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// Service is what the handlers need from the test service.
type Service interface {
	List(ctx context.Context, sc service.Scope) ([]service.ListEntry, error)
	Get(ctx context.Context, ref service.TestRef) (*models.Test, error)
	Create(ctx context.Context, sc service.Scope, in service.NewTest) (*models.Test, error)
	Update(ctx context.Context, ref service.TestRef, in service.Edit) (*models.Test, error)
}

// Handler serves the test definition routes of an app.
type Handler struct {
	svc Service
}

// New builds the handler on the service.
func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Register adds the routes to the group of an app, which must run the
// ResolveProject and ResolveApp middlewares.
func (h *Handler) Register(appAPI *gin.RouterGroup) {
	appAPI.GET("/tests", h.list)
	appAPI.POST("/test", h.create)
	appAPI.GET("/test/:test_id", h.get)
	appAPI.PUT("/test/:test_id", h.update)
}

// scope is the project and app the middlewares resolved for the request.
func scope(c *gin.Context) service.Scope {
	return service.Scope{ProjectID: httpx.Project(c).ID, AppID: httpx.App(c).ID}
}

// testRef is the test of the path inside the scope of the request. A
// non-numeric id names no test, so it is answered like an unknown one.
func testRef(c *gin.Context) (service.TestRef, error) {
	id, err := httpx.ParseID(c, "test_id")
	if err != nil {
		return service.TestRef{}, apperr.NotFound("record not found")
	}
	return service.TestRef{Scope: scope(c), TestID: id}, nil
}

func (h *Handler) list(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context(), scope(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) get(c *gin.Context) {
	ref, err := testRef(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	test, err := h.svc.Get(c.Request.Context(), ref)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, test)
}

func (h *Handler) create(c *gin.Context) {
	var in service.NewTest
	if err := httpx.BindJSON(c, &in); err != nil {
		httpx.Fail(c, err)
		return
	}
	test, err := h.svc.Create(c.Request.Context(), scope(c), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, test)
}

func (h *Handler) update(c *gin.Context) {
	var in service.Edit
	if err := httpx.BindJSON(c, &in); err != nil {
		httpx.Fail(c, err)
		return
	}
	ref, err := testRef(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	test, err := h.svc.Update(c.Request.Context(), ref, in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, test)
}
