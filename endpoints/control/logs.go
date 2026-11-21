package control

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fsuhrau/automationhub/device"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	defaultLogLines = 200
	maxLogLines     = 5000
)

// logLines reads the lines parameter; anything unreadable or outside
// 1..5000 means the default of 200.
func logLines(c *gin.Context) int {
	lines, _ := strconv.Atoi(c.DefaultQuery("lines", strconv.Itoa(defaultLogLines)))
	if lines <= 0 || lines > maxLogLines {
		return defaultLogLines
	}
	return lines
}

func (h *Handler) logs(c *gin.Context, dev device.Device) {
	result, err := h.svc.Logs(dev, logLines(c))
	if err != nil {
		failErr(c, err)
		return
	}
	if result == nil {
		result = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"lines": result})
}

// logStream follows the device log as server-sent events: "lines" carries a
// JSON array of new lines, "failure" a read error (following goes on).
// Watching needs no lease, so it works while a test run holds the device.
func (h *Handler) logStream(c *gin.Context, dev device.Device) {
	if _, ok := dev.(device.LogReader); !ok {
		failErr(c, device.ErrUnsupported)
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-store")
	// keep reverse proxies from buffering the stream
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.Writer.Flush()

	s := &eventSender{c: c}
	err := h.svc.FollowLogs(c.Request.Context(), dev, logLines(c), s.lines, s.failure)
	if err != nil {
		logrus.Debugf("log stream of %s ended: %v", dev.DeviceID(), err)
	}
}

// eventSender writes the log events of one stream.
type eventSender struct {
	c           *gin.Context
	lastFailure string
}

func (s *eventSender) lines(lines []string) error {
	s.lastFailure = ""
	return s.send("lines", lines)
}

// failure says each failure once, not every second.
func (s *eventSender) failure(err error) error {
	if err.Error() == s.lastFailure {
		return nil
	}
	s.lastFailure = err.Error()
	return s.send("failure", map[string]string{"message": err.Error()})
}

func (s *eventSender) send(event string, payload any) error {
	// the payloads are a string slice and a string map, which always marshal
	data, _ := json.Marshal(payload)
	if _, err := fmt.Fprintf(s.c.Writer, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	s.c.Writer.Flush()
	return nil
}
