package hub

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func newTestService(t *testing.T, auth config.Auth) *Service {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetOutput(&strings.Builder{})
	s := NewService(logger, net.IPv4(127, 0, 0, 1), &testkit.FakeDeviceManager{}, config.Service{Auth: auth}, nil, testkit.DB(t))
	s.authenticatedRouter.GET("/private", func(c *gin.Context) { c.String(http.StatusOK, "in") })
	return s
}

func request(s *Service, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.publicRouter.ServeHTTP(w, req)
	return w
}

func TestNewServiceWithPasswordAuthServesTheAuthRoutes(t *testing.T) {
	s := newTestService(t, config.Auth{Password: &config.Password{Secret: "test-secret-test-secret-test-sec"}})

	if w := request(s, "GET", "/auth/session", ""); w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), `"method":"password"`) {
		t.Errorf("session: %d %s", w.Code, w.Body)
	}
	if w := request(s, "GET", "/private", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("guarded route without login: %d", w.Code)
	}
	w := request(s, "POST", "/auth/register", `{"name":"Ada","email":"ada@example.com","password":"long enough"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("register: %d %s", w.Code, w.Body)
	}
	req := httptest.NewRequest("GET", "/private", nil)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	out := httptest.NewRecorder()
	s.publicRouter.ServeHTTP(out, req)
	if out.Code != http.StatusOK {
		t.Errorf("guarded route after login: %d", out.Code)
	}
}

func TestNewServiceWithoutAuthIsOpen(t *testing.T) {
	s := newTestService(t, config.Auth{})

	if w := request(s, "GET", "/auth/session", ""); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Errorf("session: %d %s", w.Code, w.Body)
	}
	if w := request(s, "GET", "/private", ""); w.Code != http.StatusOK {
		t.Errorf("open route: %d", w.Code)
	}
}
