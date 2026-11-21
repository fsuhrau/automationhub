// Package devices is the HTTP adapter of the device use cases: it reads the
// request, calls one service method and writes the result. Errors go
// through httpx.Fail, so the status code comes from the kind of the error.
package devices

import (
	"context"
	"net/http"
	"strconv"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// Service is what the handlers call; *service/device.Service implements it.
type Service interface {
	List(ctx context.Context, project *models.Project, platform *models.PlatformType) ([]models.Device, error)
	Get(ctx context.Context, project *models.Project, id string) (*models.Device, error)
	Delete(ctx context.Context, project *models.Project, id string) error
	Unlock(ctx context.Context, project *models.Project, id string) (*models.Device, error)
	Update(ctx context.Context, project *models.Project, req svc.Update) (*models.Device, error)
	RunTests(ctx context.Context, project *models.Project, req svc.TestRequest) error
	Protocols(ctx context.Context, project *models.Project, q svc.ProtocolQuery) ([]models.TestProtocol, error)
	ControlDevice(ctx context.Context, project *models.Project, id string) (device.Device, error)
	ControlApp(ctx context.Context, project *models.Project, binaryID uint) (*app.Parameter, error)
}

// Handler serves the device routes of a project.
type Handler struct {
	svc Service
}

// New returns a Handler on the service.
func New(service Service) *Handler { return &Handler{svc: service} }

// Register adds the device routes to a project group, which has to run the
// middleware that resolves the project.
func (h *Handler) Register(project *gin.RouterGroup) {
	project.GET("/devices", h.list)
	project.GET("/device/:device_id", h.get)
	project.PUT("/device/:device_id", h.update)
	project.DELETE("/device/:device_id", h.delete)
	project.POST("/device/:device_id/unlock", h.unlock)
	project.POST("/device/:device_id/tests", h.runTests)
}

func (h *Handler) list(c *gin.Context) {
	devices, err := h.svc.List(c.Request.Context(), httpx.Project(c), platformFilter(c.Query("platform")))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, devices)
}

// platformFilter reads the platform query; none means all. A value that is
// not a number selects platform 0, as it always did.
func platformFilter(query string) *models.PlatformType {
	if query == "" {
		return nil
	}
	n, _ := strconv.ParseInt(query, 10, 64)
	platform := models.PlatformType(n)
	return &platform
}

func (h *Handler) get(c *gin.Context) {
	d, err := h.svc.Get(c.Request.Context(), httpx.Project(c), c.Param("device_id"))
	respond(c, d, err)
}

func (h *Handler) unlock(c *gin.Context) {
	d, err := h.svc.Unlock(c.Request.Context(), httpx.Project(c), c.Param("device_id"))
	respond(c, d, err)
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.Project(c), c.Param("device_id")); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) update(c *gin.Context) {
	var req svc.Update
	if err := httpx.BindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	req.DeviceID = c.Param("device_id")
	d, err := h.svc.Update(c.Request.Context(), httpx.Project(c), req)
	respond(c, d, err)
}

// respond writes the device as JSON, or the error.
func respond(c *gin.Context, d *models.Device, err error) {
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}
