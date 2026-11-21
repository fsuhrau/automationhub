package hub

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func newTestService(t *testing.T, cfg config.Service, auth Auth) *Service {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetOutput(&strings.Builder{})
	s := NewService(Deps{Logger: logger, Devices: &testkit.FakeDeviceManager{}, Config: cfg, Auth: auth})
	s.authenticatedRouter.GET("/private", func(c *gin.Context) { c.String(http.StatusOK, "in") })
	return s
}

func request(s *Service, method, path string, header http.Header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range header {
		req.Header[k] = v
	}
	w := httptest.NewRecorder()
	s.publicRouter.ServeHTTP(w, req)
	return w
}

func TestNewServiceWithoutAuthIsOpen(t *testing.T) {
	s := newTestService(t, config.Service{}, Auth{})

	if w := request(s, "GET", "/auth/session", nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Errorf("session: %d %s", w.Code, w.Body)
	}
	if w := request(s, "GET", "/private", nil); w.Code != http.StatusOK {
		t.Errorf("open route: %d", w.Code)
	}
}

func TestNewServiceMountsTheAuth(t *testing.T) {
	var sessions int
	auth := Auth{
		Session: func(c *gin.Context) { sessions++ },
		Guard: func(c *gin.Context) {
			if c.GetHeader("X-Login") == "" {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
		},
		Routes: func(r *gin.Engine) { r.GET("/auth/login", func(c *gin.Context) { c.String(http.StatusOK, "login") }) },
	}
	s := newTestService(t, config.Service{}, auth)

	if w := request(s, "GET", "/auth/login", nil); w.Code != http.StatusOK || w.Body.String() != "login" {
		t.Errorf("login route: %d %s", w.Code, w.Body)
	}
	if w := request(s, "GET", "/private", nil); w.Code != http.StatusUnauthorized {
		t.Errorf("guarded route without login: %d", w.Code)
	}
	if w := request(s, "GET", "/private", http.Header{"X-Login": {"1"}}); w.Code != http.StatusOK {
		t.Errorf("guarded route after login: %d", w.Code)
	}
	if w := request(s, "GET", "/auth/session", nil); w.Code != http.StatusNotFound || sessions == 0 {
		t.Errorf("the open session route must not be mounted: %d, sessions %d", w.Code, sessions)
	}
}

func TestNewServiceAnswersPreflightsOfTheConfiguredOrigins(t *testing.T) {
	s := newTestService(t, config.Service{Cors: []string{"http://ui.test"}}, Auth{})
	header := http.Header{"Origin": {"http://ui.test"}, "Access-Control-Request-Method": {"GET"}}

	w := request(s, http.MethodOptions, "/private", header)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://ui.test" {
		t.Errorf("allowed origin %q (status %d)", got, w.Code)
	}
}

func TestNewServiceAppliesTheConfiguredLogLevel(t *testing.T) {
	tests := map[string]struct {
		level string
		want  logrus.Level
	}{
		"debug":   {"debug", logrus.DebugLevel},
		"warn":    {"warn", logrus.WarnLevel},
		"empty":   {"", logrus.InfoLevel},
		"invalid": {"loud", logrus.InfoLevel},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			logger := logrus.New()
			logger.SetOutput(&strings.Builder{})
			NewService(Deps{Logger: logger, Devices: &testkit.FakeDeviceManager{}, Config: config.Service{LogLevel: tc.level}})
			if logger.GetLevel() != tc.want {
				t.Errorf("level = %v, want %v", logger.GetLevel(), tc.want)
			}
		})
	}
}

func TestRegisterHooksKeepsOnlySlackHooks(t *testing.T) {
	s := newTestService(t, config.Service{}, Auth{})

	s.RegisterHooks([]config.Hook{{Provider: "slack"}, {Provider: "unknown"}})

	if len(s.hooks) != 1 {
		t.Errorf("hooks = %d, want 1", len(s.hooks))
	}
}

func TestHandlerServesTheRoutesOfTheService(t *testing.T) {
	s := newTestService(t, config.Service{}, Auth{})

	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/private", nil))

	if w.Code != http.StatusOK || w.Body.String() != "in" {
		t.Errorf("%d %s", w.Code, w.Body)
	}
}
