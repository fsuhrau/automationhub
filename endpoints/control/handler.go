// Package control exposes the live device control API under
// /api/:project_id/device/:device_id/...
package control

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/gin-gonic/gin"
)

// LeaseHeader carries the lease token for control calls.
const LeaseHeader = "X-Device-Lease"

// DeviceResolver resolves the device addressed by the request. It writes the
// error response itself and returns false when the device can't be used.
type DeviceResolver func(c *gin.Context) (device.Device, bool)

// AppResolver returns the app parameters for an app bundle of the project.
type AppResolver func(c *gin.Context, binaryID uint) (*app.Parameter, error)

// OwnerResolver names the caller, e.g. the user or access token name.
type OwnerResolver func(c *gin.Context) string

type Handler struct {
	svc           *hubcontrol.Service
	resolveDevice DeviceResolver
	resolveApp    AppResolver
	owner         OwnerResolver
	nodeURL       string
	// allowedOrigins may open the video websocket in addition to same-host pages
	allowedOrigins []string
}

//nolint:revive // the signature is the wiring contract of the api
func NewHandler(svc *hubcontrol.Service, resolveDevice DeviceResolver, resolveApp AppResolver, owner OwnerResolver, nodeURL string) *Handler {
	return &Handler{svc: svc, resolveDevice: resolveDevice, resolveApp: resolveApp, owner: owner, nodeURL: nodeURL}
}

// WithAllowedOrigins lets pages from these origins (e.g. the configured CORS
// origins) open the video websocket. Same-host pages are always allowed.
func (h *Handler) WithAllowedOrigins(origins []string) *Handler {
	h.allowedOrigins = origins
	return h
}

// Register adds the routes to a project scoped group.
func (h *Handler) Register(g *gin.RouterGroup) {
	d := g.Group("/device/:device_id")
	d.GET("/lock", h.withDevice(h.getLock))
	d.POST("/lock", h.withDevice(h.lock))
	d.DELETE("/lock", h.withDevice(h.release))

	d.GET("/screenshot", h.withDevice(h.screenshot))
	d.GET("/stream", h.withDevice(h.stream))
	d.GET("/video", h.withDevice(h.video))
	d.GET("/hierarchy", h.withDevice(h.hierarchy))
	d.GET("/logs", h.withDevice(h.logs))
	d.GET("/logs/stream", h.withDevice(h.logStream))

	d.POST("/tap", h.withDevice(h.tap))
	d.POST("/swipe", h.withDevice(h.swipe))
	d.POST("/text", h.withDevice(h.text))
	d.POST("/key", h.withDevice(h.key))
	d.POST("/element/touch", h.withDevice(h.elementTouch))
	d.POST("/element/value", h.withDevice(h.elementValue))

	d.POST("/app/install", h.withDevice(h.appInstall))
	d.POST("/app/start", h.withDevice(h.appStart))
	d.POST("/app/stop", h.withDevice(h.appStop))
}

func (h *Handler) withDevice(fn func(*gin.Context, device.Device)) gin.HandlerFunc {
	return func(c *gin.Context) {
		dev, ok := h.resolveDevice(c)
		if !ok {
			return
		}
		fn(c, dev)
	}
}

func lease(c *gin.Context) string {
	return c.GetHeader(LeaseHeader)
}

// target is the device of the request with the lease token it carries.
func target(c *gin.Context, dev device.Device) hubcontrol.Target {
	return hubcontrol.Target{Device: dev, Token: lease(c)}
}
