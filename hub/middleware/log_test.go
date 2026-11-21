package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func logLines(t *testing.T, clock Clock, pre gin.HandlerFunc, h gin.HandlerFunc) []string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	var out bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&out)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})
	r := gin.New()
	r.Use(pre, LoggerWithClock(logger, clock))
	r.GET("/ping", h)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.7:1234"
	r.ServeHTTP(httptest.NewRecorder(), req)
	return strings.Split(strings.TrimSpace(out.String()), "\n")
}

func noop(*gin.Context) {}

func TestLoggerLogsStartAndResultByStatus(t *testing.T) {
	cases := []struct {
		status    int
		wantLevel string
	}{{http.StatusOK, "level=info"}, {http.StatusNotFound, "level=warning"}, {http.StatusBadGateway, "level=error"}}
	for _, c := range cases {
		clock := testkit.NewFakeClock(time.Unix(0, 0))
		lines := logLines(t, clock, noop, func(ctx *gin.Context) {
			clock.Advance(1500 * time.Microsecond)
			ctx.String(c.status, "abc")
		})
		if len(lines) != 2 {
			t.Fatalf("status %d: lines %q", c.status, lines)
		}
		if !strings.Contains(lines[0], `level=info msg="[GET] \"/ping\"" address=10.0.0.7`) {
			t.Errorf("start line %q", lines[0])
		}
		want := []string{c.wantLevel, `msg="[GET] \"/ping\" 3 (2ms)"`, "latency=2", "address=10.0.0.7"}
		for _, w := range want {
			if !strings.Contains(lines[1], w) {
				t.Errorf("status %d: line %q misses %q", c.status, lines[1], w)
			}
		}
	}
}

func TestLoggerReportsPrivateErrors(t *testing.T) {
	failing := func(c *gin.Context) { _ = c.Error(http.ErrAbortHandler) }
	lines := logLines(t, systemClock{}, failing, noop)
	if len(lines) != 2 {
		t.Fatalf("lines %q", lines)
	}
	for _, l := range lines {
		if !strings.Contains(l, "level=error") || !strings.Contains(l, "http: abort Handler") {
			t.Errorf("line %q", l)
		}
	}
}

func TestLoggerUsesSystemClock(t *testing.T) {
	var out bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&out)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Logger(logger))
	r.GET("/ping", noop)
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ping", nil))
	if !strings.Contains(out.String(), "/ping") {
		t.Errorf("log %q", out.String())
	}
}

func TestLatencyMsRoundsUp(t *testing.T) {
	cases := map[time.Duration]int{0: 0, time.Nanosecond: 1, time.Millisecond: 1, 1001 * time.Microsecond: 2}
	for d, want := range cases {
		if got := latencyMs(d); got != want {
			t.Errorf("latencyMs(%v) = %d, want %d", d, got, want)
		}
	}
}
