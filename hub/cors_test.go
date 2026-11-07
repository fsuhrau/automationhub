package hub

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func corsStatus(t *testing.T, origins []string, host, origin string) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(cors.New(corsConfig(origins))) // panics on an invalid config
	r.GET("/api/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "http://"+host+"/api/ping", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestCorsWithoutConfiguredOrigins(t *testing.T) {
	cases := []struct {
		origin string
		want   int
	}{
		{"", http.StatusOK},                            // not a browser CORS request
		{"http://hub.example.com:8002", http.StatusOK}, // same origin: the embedded web UI
		{"http://localhost:3000", http.StatusOK},       // web UI dev server
		{"http://127.0.0.1:5173", http.StatusOK},
		{"http://[::1]:3000", http.StatusOK},
		{"https://evil.example.com", http.StatusForbidden},
		{"http://localhost.evil.example.com", http.StatusForbidden},
		{"file://localhost", http.StatusForbidden},
		{"null", http.StatusForbidden},
	}
	for _, c := range cases {
		if got := corsStatus(t, nil, "hub.example.com:8002", c.origin); got != c.want {
			t.Errorf("origin %q: status %d, want %d", c.origin, got, c.want)
		}
	}
}

func TestCorsWithConfiguredOrigins(t *testing.T) {
	origins := []string{"https://qa.example.com"}
	if got := corsStatus(t, origins, "hub.example.com:8002", "https://qa.example.com"); got != http.StatusOK {
		t.Errorf("listed origin: status %d, want 200", got)
	}
	// The list replaces the default, so localhost is no longer allowed.
	if got := corsStatus(t, origins, "hub.example.com:8002", "http://localhost:3000"); got != http.StatusForbidden {
		t.Errorf("unlisted localhost: status %d, want 403", got)
	}
}
