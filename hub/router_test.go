package hub

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/hub/session"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestNewRouterRecoversPanicsAndLogsRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var out bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&out)
	r := newRouter(logger)
	r.GET("/boom", func(*gin.Context) { panic("boom") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", w.Code)
	}
	if !strings.Contains(out.String(), "prefix=service") || !strings.Contains(out.String(), "/boom") {
		t.Errorf("request not logged: %q", out.String())
	}
}

func TestNewSessionManagerIsASessionManager(t *testing.T) {
	m := NewSessionManager(logrus.New(), nil)
	if _, ok := any(m).(*session.Manager); !ok || m == nil || len(m.GetSessions()) != 0 {
		t.Error("expected an empty manager")
	}
}
