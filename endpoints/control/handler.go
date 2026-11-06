// Package control exposes the live device control API under
// /api/:project_id/device/:device_id/...
package control

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/fsuhrau/automationhub/utils/sync"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// LeaseHeader carries the lease token for control calls.
const LeaseHeader = "X-Device-Lease"

const (
	defaultStreamFPS = 2
	maxStreamFPS     = 5
	streamBoundary   = "frame"
	maxStreamBackoff = 10 * time.Second

	defaultVideoMaxSize = 1280
	maxVideoMaxSize     = 4096
	defaultVideoFPS     = 30
	maxVideoFPS         = 60
	defaultVideoBitRate = 4_000_000
	minVideoBitRate     = 250_000
	maxVideoBitRate     = 20_000_000
)

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

type ErrorResponse struct {
	Message string `json:"message"`
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

// statusFor maps service errors to HTTP status codes.
func statusFor(err error) int {
	switch {
	case errors.Is(err, hubcontrol.ErrLeaseRequired):
		return http.StatusPreconditionRequired
	case errors.Is(err, hubcontrol.ErrLeaseInvalid), errors.Is(err, hubcontrol.ErrDeviceBusy), errors.Is(err, hubcontrol.ErrAppNotConnected):
		return http.StatusConflict
	case errors.Is(err, device.ErrUnsupported):
		return http.StatusNotImplemented
	case errors.Is(err, sync.TimeoutError):
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}

func fail(c *gin.Context, status int, err error) {
	c.AbortWithStatusJSON(status, ErrorResponse{Message: err.Error()})
}

func failErr(c *gin.Context, err error) {
	fail(c, statusFor(err), err)
}

func lease(c *gin.Context) string {
	return c.GetHeader(LeaseHeader)
}

func bind(c *gin.Context, v any) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		fail(c, http.StatusBadRequest, err)
		return false
	}
	return true
}

// --- lock ---

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

// --- read ---

func (h *Handler) screenshot(c *gin.Context, dev device.Device) {
	var (
		data          []byte
		width, height int
		err           error
	)
	if c.Query("source") == "app" {
		data, width, height, err = h.svc.AppScreenshot(c.Request.Context(), dev)
	} else {
		data, width, height, err = h.svc.Screenshot(dev)
	}
	if err != nil {
		failErr(c, err)
		return
	}
	c.Header("X-Screen-Width", strconv.Itoa(width))
	c.Header("X-Screen-Height", strconv.Itoa(height))
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, http.DetectContentType(data), data)
}

// stream sends native screenshots as a multipart MJPEG stream, usable as an
// <img> source. Frame size headers carry the device resolution.
func (h *Handler) stream(c *gin.Context, dev device.Device) {
	fps, _ := strconv.Atoi(c.DefaultQuery("fps", strconv.Itoa(defaultStreamFPS)))
	if fps < 1 {
		fps = 1
	}
	if fps > maxStreamFPS {
		fps = maxStreamFPS
	}
	interval := time.Second / time.Duration(fps)

	c.Header("Content-Type", "multipart/x-mixed-replace; boundary="+streamBoundary)
	c.Header("Cache-Control", "no-store")
	c.Status(http.StatusOK)

	ctx := c.Request.Context()
	// on errors (e.g. the device was unplugged) back off instead of retrying
	// at the frame rate
	wait := interval
	for {
		data, w, hgt, err := h.svc.Screenshot(dev)
		if err != nil {
			wait = min(wait*2, maxStreamBackoff)
		} else {
			wait = interval
			frame, err := toJPEG(data)
			if err == nil {
				_, err = fmt.Fprintf(c.Writer, "--%s\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\nX-Screen-Width: %d\r\nX-Screen-Height: %d\r\n\r\n", streamBoundary, len(frame), w, hgt)
				if err == nil {
					_, err = c.Writer.Write(append(frame, '\r', '\n'))
				}
				if err != nil {
					return
				}
				c.Writer.Flush()
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

// video streams the screen as H.264 over a websocket; see package
// videostream for the messages. Devices without video support get an error
// message with "unsupported" set, so the client can fall back to /stream.
// Watching needs no lease; "input" messages must carry the lease token.
func (h *Handler) video(c *gin.Context, dev device.Device) {
	opts := device.VideoOptions{
		MaxSize: queryInt(c, "maxSize", defaultVideoMaxSize, 0, maxVideoMaxSize),
		MaxFPS:  queryInt(c, "maxFps", defaultVideoFPS, 1, maxVideoFPS),
		BitRate: queryInt(c, "bitRate", defaultVideoBitRate, minVideoBitRate, maxVideoBitRate),
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  4 << 10,
		WriteBufferSize: 256 << 10,
		CheckOrigin:     h.checkOrigin,
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// the upgrader already wrote the error response
		return
	}
	defer conn.Close()

	// the connection is hijacked, so the request context doesn't notice a
	// client that goes away; Pump watches the socket for that.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := h.svc.Video(ctx, dev, opts)
	if err != nil {
		_ = videostream.WriteError(conn, err)
	} else if err := videostream.Pump(ctx, conn, stream, h.videoInput(dev, stream)); err != nil {
		logrus.Debugf("video stream of %s ended: %v", dev.DeviceID(), err)
	}
	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
}

func (h *Handler) videoInput(dev device.Device, stream device.VideoStream) videostream.InputHandler {
	return func(m videostream.Message, reply func(videostream.Message)) {
		if m.Type != videostream.TypeInput || m.Input == nil {
			return
		}
		if err := h.svc.InjectInput(dev, m.Lease, stream, *m.Input); err != nil {
			reply(videostream.Message{
				Type:        videostream.TypeInputError,
				Message:     err.Error(),
				Unsupported: errors.Is(err, device.ErrUnsupported),
			})
		}
	}
}

// checkOrigin blocks cross-site websocket requests, which would otherwise
// carry the user's session cookie.
func (h *Handler) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}
	for _, allowed := range h.allowedOrigins {
		if allowed == "*" || strings.EqualFold(strings.TrimSuffix(allowed, "/"), origin) {
			return true
		}
	}
	return false
}

func queryInt(c *gin.Context, key string, def, lo, hi int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return def
	}
	return max(lo, min(v, hi))
}

func toJPEG(data []byte) ([]byte, error) {
	if http.DetectContentType(data) == "image/jpeg" {
		return data, nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 70}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h *Handler) hierarchy(c *gin.Context, dev device.Device) {
	xml, err := h.svc.Hierarchy(c.Request.Context(), dev, c.Query("pretty") != "false")
	if err != nil {
		failErr(c, err)
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", xml)
}

func (h *Handler) logs(c *gin.Context, dev device.Device) {
	lines, _ := strconv.Atoi(c.DefaultQuery("lines", "200"))
	if lines <= 0 || lines > 5000 {
		lines = 200
	}
	result, err := h.svc.Logs(dev, lines)
	if err != nil {
		failErr(c, err)
		return
	}
	if result == nil {
		result = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"lines": result})
}

// --- control ---

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
	var err error
	switch {
	case !req.ElementRef.IsZero():
		err = h.svc.TouchElement(c.Request.Context(), dev, lease(c), req.ElementRef)
	case req.X != nil && req.Y != nil:
		err = h.svc.Tap(dev, lease(c), *req.X, *req.Y)
	default:
		fail(c, http.StatusBadRequest, fmt.Errorf("x and y, or xpath/elementId are required"))
		return
	}
	if err != nil {
		failErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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
	if err := h.svc.Swipe(dev, lease(c), req.FromX, req.FromY, req.ToX, req.ToY, time.Duration(req.DurationMs)*time.Millisecond); err != nil {
		failErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type textRequest struct {
	Text string `json:"text" binding:"required"`
}

func (h *Handler) text(c *gin.Context, dev device.Device) {
	var req textRequest
	if !bind(c, &req) {
		return
	}
	if err := h.svc.InputText(dev, lease(c), req.Text); err != nil {
		failErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type keyRequest struct {
	Key string `json:"key" binding:"required"`
}

func (h *Handler) key(c *gin.Context, dev device.Device) {
	var req keyRequest
	if !bind(c, &req) {
		return
	}
	if err := h.svc.PressKey(dev, lease(c), req.Key); err != nil {
		failErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) elementTouch(c *gin.Context, dev device.Device) {
	var req hubcontrol.ElementRef
	if !bind(c, &req) {
		return
	}
	if err := h.svc.TouchElement(c.Request.Context(), dev, lease(c), req); err != nil {
		failErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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
		req.Attr = "text"
	}
	if err := h.svc.SetElementValue(c.Request.Context(), dev, lease(c), req.ElementRef, req.Attr, req.Value); err != nil {
		failErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- app lifecycle ---

type appRequest struct {
	BinaryID  uint `json:"binaryId" binding:"required"`
	ClearData bool `json:"clearData"`
}

func (h *Handler) appParams(c *gin.Context) (*app.Parameter, *appRequest, bool) {
	var req appRequest
	if !bind(c, &req) {
		return nil, nil, false
	}
	params, err := h.resolveApp(c, req.BinaryID)
	if err != nil {
		fail(c, http.StatusNotFound, err)
		return nil, nil, false
	}
	return params, &req, true
}

func (h *Handler) appInstall(c *gin.Context, dev device.Device) {
	params, _, ok := h.appParams(c)
	if !ok {
		return
	}
	if err := h.svc.InstallApp(c.Request.Context(), dev, lease(c), params); err != nil {
		failErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) appStart(c *gin.Context, dev device.Device) {
	params, _, ok := h.appParams(c)
	if !ok {
		return
	}
	sessionID := uuid.NewString()
	if err := h.svc.StartApp(dev, lease(c), params, sessionID, h.nodeURL); err != nil {
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
	if err := h.svc.StopApp(dev, lease(c), params, req.ClearData); err != nil {
		failErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
