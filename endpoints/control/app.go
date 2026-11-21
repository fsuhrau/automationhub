package control

import (
	"net/http"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type appRequest struct {
	BinaryID  uint `json:"binaryId" binding:"required"`
	ClearData bool `json:"clearData"`
}

// appParams reads the request and resolves its app bundle; it answers the
// error itself and reports false when the call can't go on.
func (h *Handler) appParams(c *gin.Context) (*app.Parameter, *appRequest, bool) {
	var req appRequest
	if !bind(c, &req) {
		return nil, nil, false
	}
	params, err := h.resolveApp(c, req.BinaryID)
	if err != nil {
		httpx.Fail(c, apperr.NotFound("%w", err))
		return nil, nil, false
	}
	return params, &req, true
}

func (h *Handler) appInstall(c *gin.Context, dev device.Device) {
	params, _, ok := h.appParams(c)
	if !ok {
		return
	}
	done(c, h.svc.InstallApp(c.Request.Context(), target(c, dev), params))
}

func (h *Handler) appStart(c *gin.Context, dev device.Device) {
	params, _, ok := h.appParams(c)
	if !ok {
		return
	}
	sessionID := uuid.NewString()
	if err := h.svc.StartApp(target(c, dev), params, hubcontrol.Launch{SessionID: sessionID, NodeURL: h.nodeURL}); err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessionId": sessionID})
}

func (h *Handler) appStop(c *gin.Context, dev device.Device) {
	params, req, ok := h.appParams(c)
	if !ok {
		return
	}
	done(c, h.svc.StopApp(target(c, dev), params, req.ClearData))
}
