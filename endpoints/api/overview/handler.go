// Package overview holds the HTTP handlers of the project overview and
// statistics. They call the overview service and write the answer.
package overview

import (
	"context"
	"net/http"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
	svc "github.com/fsuhrau/automationhub/service/overview"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// Service is what the handlers need from the overview service.
type Service interface {
	Get(ctx context.Context, projectID uint) (*svc.Overview, error)
	Stats(ctx context.Context, project *models.Project) (*svc.Stats, error)
}

// Handler serves the overview and statistics routes of a project.
type Handler struct {
	svc Service
}

// New builds the handler on the service.
func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterOverview adds the overview to the group of a project, which must
// run the ResolveProject middleware.
func (h *Handler) RegisterOverview(projectAPI gin.IRoutes) {
	projectAPI.GET("/overview", h.overview)
}

// RegisterStats adds the statistics to the group of a project, which must
// run the ResolveProject middleware.
func (h *Handler) RegisterStats(projectAPI gin.IRoutes) {
	projectAPI.GET("/stats", h.stats)
}

func (h *Handler) overview(c *gin.Context) {
	o, err := h.svc.Get(c.Request.Context(), httpx.Project(c).ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *Handler) stats(c *gin.Context) {
	stats, err := h.svc.Stats(c.Request.Context(), httpx.Project(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}
