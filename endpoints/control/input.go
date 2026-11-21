package control

import (
	"net/http"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/gin-gonic/gin"
)

const defaultElementAttr = "text"

// done answers 204 for a successful input call and the error otherwise.
func done(c *gin.Context, err error) {
	if err != nil {
		failErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type tapRequest struct {
	X *int `json:"x"`
	Y *int `json:"y"`
	hubcontrol.ElementRef
}

// tap taps a point, or touches an app element when xpath/elementId is given.
func (h *Handler) tap(c *gin.Context, dev device.Device) {
	var req tapRequest
	if !bind(c, &req) {
		return
	}
	switch {
	case !req.IsZero():
		done(c, h.svc.TouchElement(c.Request.Context(), target(c, dev), req.ElementRef))
	case req.X != nil && req.Y != nil:
		done(c, h.svc.Tap(target(c, dev), *req.X, *req.Y))
	default:
		httpx.Fail(c, apperr.Invalid("x and y, or xpath/elementId are required"))
	}
}

type swipeRequest struct {
	FromX      int `json:"fromX"`
	FromY      int `json:"fromY"`
	ToX        int `json:"toX"`
	ToY        int `json:"toY"`
	DurationMs int `json:"durationMs"`
}

func (h *Handler) swipe(c *gin.Context, dev device.Device) {
	var req swipeRequest
	if !bind(c, &req) {
		return
	}
	gesture := hubcontrol.SwipeGesture{
		FromX: req.FromX, FromY: req.FromY, ToX: req.ToX, ToY: req.ToY,
		Duration: time.Duration(req.DurationMs) * time.Millisecond,
	}
	done(c, h.svc.Swipe(target(c, dev), gesture))
}

type textRequest struct {
	Text string `json:"text" binding:"required"`
}

func (h *Handler) text(c *gin.Context, dev device.Device) {
	var req textRequest
	if !bind(c, &req) {
		return
	}
	done(c, h.svc.InputText(target(c, dev), req.Text))
}

type keyRequest struct {
	Key string `json:"key" binding:"required"`
}

func (h *Handler) key(c *gin.Context, dev device.Device) {
	var req keyRequest
	if !bind(c, &req) {
		return
	}
	done(c, h.svc.PressKey(target(c, dev), req.Key))
}

func (h *Handler) elementTouch(c *gin.Context, dev device.Device) {
	var req hubcontrol.ElementRef
	if !bind(c, &req) {
		return
	}
	done(c, h.svc.TouchElement(c.Request.Context(), target(c, dev), req))
}

type elementValueRequest struct {
	hubcontrol.ElementRef
	Attr  string `json:"attr"`
	Value string `json:"value"`
}

func (h *Handler) elementValue(c *gin.Context, dev device.Device) {
	var req elementValueRequest
	if !bind(c, &req) {
		return
	}
	if req.Attr == "" {
		req.Attr = defaultElementAttr
	}
	done(c, h.svc.SetElementValue(c.Request.Context(), target(c, dev), hubcontrol.ElementChange{Ref: req.ElementRef, Attr: req.Attr, Value: req.Value}))
}
