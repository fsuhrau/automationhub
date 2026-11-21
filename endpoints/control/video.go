package control

import (
	"context"
	"errors"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	defaultVideoMaxSize = 1280
	maxVideoMaxSize     = 4096
	defaultVideoFPS     = 30
	maxVideoFPS         = 60
	defaultVideoBitRate = 4_000_000
	minVideoBitRate     = 250_000
	maxVideoBitRate     = 20_000_000
)

func videoOptions(c *gin.Context) device.VideoOptions {
	return device.VideoOptions{
		MaxSize: clampQuery(c, "maxSize", bounds{defaultVideoMaxSize, 0, maxVideoMaxSize}),
		MaxFPS:  clampQuery(c, "maxFps", bounds{defaultVideoFPS, 1, maxVideoFPS}),
		BitRate: clampQuery(c, "bitRate", bounds{defaultVideoBitRate, minVideoBitRate, maxVideoBitRate}),
	}
}

// video streams the screen as H.264 over a websocket; see package
// videostream for the messages. Devices without video support get an error
// message with "unsupported" set, so the client can fall back to /stream.
// Watching needs no lease; "input" messages must carry the lease token.
func (h *Handler) video(c *gin.Context, dev device.Device) {
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
	defer func() { _ = conn.Close() }()
	h.pumpVideo(conn, dev, videoOptions(c))
	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
}

func (h *Handler) pumpVideo(conn *websocket.Conn, dev device.Device, opts device.VideoOptions) {
	// the connection is hijacked, so the request context doesn't notice a
	// client that goes away; Pump watches the socket for that.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := h.svc.Video(ctx, dev, opts)
	if err != nil {
		_ = videostream.WriteError(conn, err)
		return
	}
	if err := videostream.Pump(ctx, conn, videostream.Feed{Stream: stream, Input: h.videoInput(dev, stream)}); err != nil {
		logrus.Debugf("video stream of %s ended: %v", dev.DeviceID(), err)
	}
}

func (h *Handler) videoInput(dev device.Device, stream device.VideoStream) videostream.InputHandler {
	return func(m videostream.Message, reply func(videostream.Message)) {
		if m.Type != videostream.TypeInput || m.Input == nil {
			return
		}
		if err := h.svc.InjectInput(hubcontrol.Target{Device: dev, Token: m.Lease}, stream, *m.Input); err != nil {
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
	return strings.EqualFold(u.Host, r.Host) || h.originAllowed(origin)
}

func (h *Handler) originAllowed(origin string) bool {
	for _, allowed := range h.allowedOrigins {
		if allowed == "*" || strings.EqualFold(strings.TrimSuffix(allowed, "/"), origin) {
			return true
		}
	}
	return false
}
