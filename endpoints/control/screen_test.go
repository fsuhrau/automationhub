package control

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/gin-gonic/gin"
)

func TestScreenshot(t *testing.T) {
	e := setup(t)
	rec := e.call(t, "", http.MethodGet, "1/screenshot", nil)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("unexpected response %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("X-Screen-Width") != "4" || rec.Header().Get("X-Screen-Height") != "2" {
		t.Fatal("missing screen size headers")
	}
}

func TestScreenshotFailures(t *testing.T) {
	e := setup(t)
	if rec := e.call(t, "", http.MethodGet, "1/screenshot?source=app", nil); rec.Code != http.StatusConflict {
		t.Fatalf("app screenshot without app: expected 409, got %d %s", rec.Code, rec.Body)
	}
	e.full.FailWith("GetScreenshot", errors.New("boom"))
	if rec := e.call(t, "", http.MethodGet, "1/screenshot", nil); rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestHierarchy(t *testing.T) {
	e := setup(t)
	e.full.AppConnected = true
	rec := e.call(t, "", http.MethodGet, "1/hierarchy?pretty=false", nil)
	if rec.Code != http.StatusOK || rec.Body.String() != sceneXML || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/xml") {
		t.Fatalf("hierarchy: %d %s %s", rec.Code, rec.Header().Get("Content-Type"), rec.Body)
	}
}

func TestAppNotConnected(t *testing.T) {
	e := setup(t)
	if rec := e.call(t, "", http.MethodGet, "1/hierarchy", nil); rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "no app connected") {
		t.Fatalf("expected app not connected error, got %d %s", rec.Code, rec.Body)
	}
}

func TestStreamSendsJPEGFrames(t *testing.T) {
	e := setup(t)
	srv := httptest.NewServer(e.router)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/p/device/1/stream?fps=5", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "multipart/x-mixed-replace") {
		t.Fatalf("unexpected content type %s", resp.Header.Get("Content-Type"))
	}
	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	head := string(buf[:n])
	if !strings.Contains(head, "--frame") || !strings.Contains(head, "image/jpeg") || !strings.Contains(head, "X-Screen-Width: 4") {
		t.Fatalf("no jpeg frame in stream: %q", head)
	}
}

func TestStreamBacksOffOnErrorsAndStopsWithTheClient(t *testing.T) {
	e := setup(t)
	e.full.FailWith("GetScreenshot", errors.New("unplugged"))
	ctx, cancel := context.WithCancel(context.Background())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/p/device/1/stream", nil).WithContext(ctx)
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	done := make(chan struct{})
	go func() { e.router.ServeHTTP(rec, req); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("stream did not stop when the client left")
	}
	if strings.Contains(rec.Body.String(), "--frame") {
		t.Fatalf("frames sent although the screenshot fails: %q", rec.Body)
	}
}

func TestStreamFPS(t *testing.T) {
	tests := map[string]int{"": 2, "4": 4, "99": 5, "0": 1, "-3": 1, "abc": 1}
	for q, want := range tests {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/?fps="+url.QueryEscape(q), nil)
		if q == "" {
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		}
		if got := streamFPS(c); got != want {
			t.Errorf("fps=%q: got %d, want %d", q, got, want)
		}
	}
}

func TestClampQuery(t *testing.T) {
	tests := map[string]int{"": 10, "abc": 10, "5": 5, "1": 2, "99": 20}
	for q, want := range tests {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/?k="+q, nil)
		if got := clampQuery(c, "k", bounds{10, 2, 20}); got != want {
			t.Errorf("k=%q: got %d, want %d", q, got, want)
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestWriteFrame(t *testing.T) {
	var out bytes.Buffer
	if err := writeFrame(&out, frame{testPNG(t), 4, 2}); err != nil {
		t.Fatal(err)
	}
	part := out.String()
	for _, want := range []string{"--frame\r\n", "Content-Type: image/jpeg\r\n", "X-Screen-Width: 4\r\n", "X-Screen-Height: 2\r\n\r\n"} {
		if !strings.Contains(part, want) {
			t.Fatalf("part lacks %q: %q", want, part)
		}
	}
	if !strings.HasSuffix(part, "\r\n") {
		t.Fatal("part does not end with a line break")
	}
	if err := writeFrame(failingWriter{}, frame{testPNG(t), 4, 2}); err == nil {
		t.Fatal("write error swallowed")
	}
	out.Reset()
	if err := writeFrame(&out, frame{[]byte("not an image"), 4, 2}); err != nil || out.Len() != 0 {
		t.Fatalf("undecodable screenshot must be skipped: %v %d", err, out.Len())
	}
}

func TestToJPEG(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatal(err)
	}
	if got, err := toJPEG(buf.Bytes()); err != nil || !bytes.Equal(got, buf.Bytes()) {
		t.Fatalf("a jpeg must pass unchanged: %v", err)
	}
	got, err := toJPEG(testPNG(t))
	if err != nil || http.DetectContentType(got) != "image/jpeg" {
		t.Fatalf("png not converted: %v", err)
	}
	if _, err := toJPEG([]byte("x")); err == nil {
		t.Fatal("expected a decode error")
	}
}

func TestSleep(t *testing.T) {
	if !sleep(context.Background(), time.Millisecond) {
		t.Fatal("sleep reported cancel")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if sleep(ctx, time.Hour) {
		t.Fatal("sleep ignored the cancelled context")
	}
}

func TestPumpFramesStopsWhenTheClientWriteFails(t *testing.T) {
	e := setup(t)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Writer = errWriter{c.Writer}
	h := &Handler{svc: hubcontrol.New(nil)}
	done := make(chan struct{})
	go func() { h.pumpFrames(c, e.full, time.Hour); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("frames still pumped after a write error")
	}
}
