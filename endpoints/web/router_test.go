package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/fsuhrau/automationhub/config"
	"github.com/gin-gonic/gin"
)

type brokenFS struct{}

func (brokenFS) Open(string) (http.File, error) { return nil, fs.ErrNotExist }

type unreadableFS struct{ http.FileSystem }

func (unreadableFS) Open(string) (http.File, error) { return http.Dir(".").Open("/") }

func newRouter(t *testing.T, fsys http.FileSystem) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	s := New(nil, config.Service{})
	s.filesystem = fsys
	r := gin.New()
	if err := s.RegisterRoutes(r, r.Group("/")); err != nil {
		t.Fatal(err)
	}
	return r
}

func mapFS() http.FileSystem {
	return http.FS(fstest.MapFS{
		"data/index.html":        {Data: []byte("<html>index</html>")},
		"data/manifest.json":     {Data: []byte(`{"a":1}`)},
		"data/static/js/main.js": {Data: []byte("console.log(1)")},
	})
}

func get(r http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestRoutes(t *testing.T) {
	r := newRouter(t, mapFS())
	tests := []struct {
		name, path, body string
		code             int
	}{
		{"static asset", "/static/js/main.js", "console.log(1)", 200},
		{"manifest", "/manifest.json", `{"a":1}`, 200},
		{"spa fallback", "/devices/42", "<html>index</html>", 200},
		{"root serves index", "/", "<html>index</html>", 200},
		{"missing asset", "/static/js/missing.js", "404 page not found\n", 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := get(r, tt.path)
			if rec.Code != tt.code || rec.Body.String() != tt.body {
				t.Errorf("got %d %q", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRegisterRoutes_IndexMissing(t *testing.T) {
	s := New(nil, config.Service{})
	s.filesystem = brokenFS{}
	if err := s.RegisterRoutes(gin.New(), gin.New().Group("/")); err == nil {
		t.Fatal("expected an error when index.html is missing")
	}
}

func TestRegisterRoutes_IndexUnreadable(t *testing.T) {
	s := New(nil, config.Service{})
	s.filesystem = unreadableFS{} // opens a directory, reading it fails
	if err := s.RegisterRoutes(gin.New(), gin.New().Group("/")); err == nil {
		t.Fatal("expected an error when index.html cannot be read")
	}
}

func TestNew_DefaultsToEmbeddedContent(t *testing.T) {
	f, err := New(nil, config.Service{}).filesystem.Open("/data/index.html")
	if err != nil {
		t.Skipf("frontend build not embedded: %v", err)
	}
	_ = f.Close()
}
