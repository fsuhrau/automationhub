package control

import (
	"net/http"

	"github.com/fsuhrau/automationhub/device"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/gin-gonic/gin"
)

type lockResponse struct {
	Locked bool              `json:"locked"`
	Lease  *hubcontrol.Lease `json:"lease,omitempty"`
	// LockedByTest is set when the device is locked outside of a control lease
	LockedByTest bool `json:"lockedByTest"`
}

func (h *Handler) getLock(c *gin.Context, dev device.Device) {
	resp := lockResponse{Locked: dev.IsLocked()}
	if l, ok := h.svc.Holder(dev.DeviceID()); ok {
		resp.Lease = &l
	} else if resp.Locked {
		resp.LockedByTest = true
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) lock(c *gin.Context, dev device.Device) {
	l, err := h.svc.Acquire(dev, h.owner(c))
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, l)
}

func (h *Handler) release(c *gin.Context, dev device.Device) {
	if err := h.svc.Release(dev.DeviceID(), lease(c)); err != nil {
		failErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
