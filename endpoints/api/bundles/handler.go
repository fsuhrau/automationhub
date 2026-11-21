// Package bundles is the HTTP adapter of the bundles of an app: it reads the
// request, calls the app service and writes the result.
package bundles

import (
	"context"
	"net/http"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// uploadField is the form field the build file is sent in.
const uploadField = "test_target"

// Service is what the handlers call.
type Service interface {
	Bundles(ctx context.Context, sc appsvc.Scope) ([]appsvc.BundleEntry, error)
	Bundle(ctx context.Context, ref appsvc.BundleRef) (*appsvc.BundleDetail, error)
	SetTags(ctx context.Context, ref appsvc.BundleRef, tags string) (*models.AppBinary, error)
	DeleteBundle(ctx context.Context, ref appsvc.BundleRef) error
	Archive(ctx context.Context, ref appsvc.BundleRef) (*appsvc.BundleDetail, error)
	ReadAgain(ctx context.Context, ref appsvc.BundleRef) (*appsvc.BundleDetail, error)
	Icon(ctx context.Context, ref appsvc.BundleRef) (*appsvc.Icon, error)
	Upload(ctx context.Context, sc appsvc.Scope, file appsvc.UploadFile) (*models.AppBinary, error)
}

// Handler serves the bundle routes.
type Handler struct {
	svc Service
}

// New builds the handler on the service.
func New(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds the routes below the group that resolves the app.
func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/bundles", h.list)
	r.GET("/bundle/:binary_id", detail(h.svc.Bundle))
	r.GET("/bundle/:binary_id/icon", h.icon)
	r.POST("/bundle/:binary_id/read", detail(h.svc.ReadAgain))
	r.POST("/bundle/:binary_id/archive", detail(h.svc.Archive))
	r.PUT("/bundle/:binary_id", h.setTags)
	r.DELETE("/bundle/:binary_id", h.remove)
	r.POST("/upload", h.upload)
}

func scope(c *gin.Context) appsvc.Scope {
	return appsvc.Scope{ProjectID: httpx.Project(c).ID, AppID: httpx.App(c).ID}
}

func bundleRef(c *gin.Context) (appsvc.BundleRef, error) {
	id, err := httpx.ParseID(c, "binary_id")
	return appsvc.BundleRef{Scope: scope(c), ID: id}, err
}

// detail serves the routes that answer with the bundle page.
func detail(call func(context.Context, appsvc.BundleRef) (*appsvc.BundleDetail, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		ref, err := bundleRef(c)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		found, err := call(c.Request.Context(), ref)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		c.JSON(http.StatusOK, found)
	}
}

func (h *Handler) list(c *gin.Context) {
	found, err := h.svc.Bundles(c.Request.Context(), scope(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, found)
}

func (h *Handler) icon(c *gin.Context) {
	ref, err := bundleRef(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	icon, err := h.svc.Icon(c.Request.Context(), ref)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	// a bundle's file never changes, so neither does its icon
	c.Header("Cache-Control", "private, max-age=86400")
	c.Data(http.StatusOK, icon.ContentType, icon.Data)
}

func (h *Handler) setTags(c *gin.Context) {
	ref, err := bundleRef(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req struct {
		Tags string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, apperr.Invalid("%s", err))
		return
	}
	updated, err := h.svc.SetTags(c.Request.Context(), ref, req.Tags)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) remove(c *gin.Context) {
	ref, err := bundleRef(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := h.svc.DeleteBundle(c.Request.Context(), ref); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.Status(http.StatusOK)
}
