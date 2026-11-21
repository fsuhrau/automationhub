package control

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	_ "image/png" // decodes png screenshots for the MJPEG stream
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/gin-gonic/gin"
)

const (
	defaultStreamFPS = 2
	maxStreamFPS     = 5
	streamBoundary   = "frame"
	maxStreamBackoff = 10 * time.Second
	jpegQuality      = 70
)

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

func (h *Handler) hierarchy(c *gin.Context, dev device.Device) {
	xml, err := h.svc.Hierarchy(c.Request.Context(), dev, c.Query("pretty") != "false")
	if err != nil {
		failErr(c, err)
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", xml)
}

// stream sends native screenshots as a multipart MJPEG stream, usable as an
// <img> source. Frame size headers carry the device resolution.
func (h *Handler) stream(c *gin.Context, dev device.Device) {
	fps := streamFPS(c)
	c.Header("Content-Type", "multipart/x-mixed-replace; boundary="+streamBoundary)
	c.Header("Cache-Control", "no-store")
	c.Status(http.StatusOK)
	h.pumpFrames(c, dev, time.Second/time.Duration(fps))
}

// pumpFrames writes a frame every interval until the client leaves. On
// errors (e.g. the device was unplugged) it backs off instead of retrying at
// the frame rate.
func (h *Handler) pumpFrames(c *gin.Context, dev device.Device, interval time.Duration) {
	wait := interval
	for {
		data, w, hgt, err := h.svc.Screenshot(dev)
		if err != nil {
			wait = min(wait*2, maxStreamBackoff)
		} else {
			wait = interval
			if writeFrame(c.Writer, frame{data, w, hgt}) != nil {
				return
			}
			c.Writer.Flush()
		}
		if !sleep(c.Request.Context(), wait) {
			return
		}
	}
}

// sleep waits for d and reports false when ctx ended first.
func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// frame is a screenshot with the screen size of the device.
type frame struct {
	data          []byte
	width, height int
}

// writeFrame writes one multipart part. A screenshot that can't be turned
// into a JPEG is skipped.
func writeFrame(out io.Writer, f frame) error {
	jpg, err := toJPEG(f.data)
	if err != nil {
		return nil
	}
	var part bytes.Buffer
	part.WriteString("--" + streamBoundary + "\r\nContent-Type: image/jpeg\r\nContent-Length: " + strconv.Itoa(len(jpg)))
	part.WriteString("\r\nX-Screen-Width: " + strconv.Itoa(f.width) + "\r\nX-Screen-Height: " + strconv.Itoa(f.height) + "\r\n\r\n")
	part.Write(jpg)
	part.WriteString("\r\n")
	_, err = out.Write(part.Bytes())
	return err
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
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpegQuality})
	return buf.Bytes(), err
}

// streamFPS reads the fps parameter: 2 when missing, limited to 1..5 (an
// unreadable value counts as 0 and so becomes 1).
func streamFPS(c *gin.Context) int {
	fps, _ := strconv.Atoi(c.DefaultQuery("fps", strconv.Itoa(defaultStreamFPS)))
	return max(1, min(fps, maxStreamFPS))
}

// bounds is the default of a query parameter and the range of its values.
type bounds struct{ def, lo, hi int }

// clampQuery reads the integer query parameter key: the default when it is
// missing or not a number, otherwise the value limited to the range.
// (httpx.Limit only reads "limit" and treats non-positive values differently.)
func clampQuery(c *gin.Context, key string, b bounds) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return b.def
	}
	return max(b.lo, min(v, b.hi))
}
