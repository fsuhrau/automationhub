package middleware

import (
	"bytes"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func serve(mw gin.HandlerFunc, h gin.HandlerFunc, header map[string]string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw)
	r.GET("/boom", h)
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func opErr(msg error) *net.OpError {
	return &net.OpError{Op: "write", Err: os.NewSyscallError("write", msg)}
}

func assertLog(t *testing.T, log string, want, wantNot []string) {
	t.Helper()
	for _, s := range want {
		if !strings.Contains(log, s) {
			t.Errorf("log %q misses %q", log, s)
		}
	}
	for _, s := range wantNot {
		if strings.Contains(log, s) {
			t.Errorf("log %q must not contain %q", log, s)
		}
	}
}

func TestRecoveryAnswersPanics(t *testing.T) {
	cases := []struct {
		name       string
		panicValue any
		wantLog    []string
		wantNoLog  []string
		wantErrors int
	}{
		{"error panic", errors.New("kaputt"), []string{"[Recovery]", "kaputt", "recovery_test.go"}, nil, 0},
		{"string panic", "plain text", []string{"[Recovery]", "plain text"}, nil, 0},
		{"broken pipe", opErr(syscall.EPIPE), []string{"broken pipe"}, []string{"[Recovery]"}, 1},
		{"connection reset", opErr(syscall.ECONNRESET), []string{"connection reset by peer"}, []string{"[Recovery]"}, 1},
		{"other op error", opErr(syscall.EACCES), []string{"[Recovery]"}, nil, 0},
		{"op error without syscall", &net.OpError{Err: errors.New("x")}, []string{"[Recovery]"}, nil, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			w := serve(RecoveryWithWriter(&out), func(*gin.Context) { panic(c.panicValue) }, nil)
			if w.Code != http.StatusInternalServerError || w.Body.Len() != 0 {
				t.Errorf("response %d %q, want an empty 500", w.Code, w.Body.String())
			}
			assertLog(t, out.String(), c.wantLog, c.wantNoLog)
		})
	}
}

func TestRecoveryKeepsHealthyAndAbortedRequests(t *testing.T) {
	w := serve(Recovery(), func(c *gin.Context) { c.String(http.StatusTeapot, "ok") }, nil)
	if w.Code != http.StatusTeapot || w.Body.String() != "ok" {
		t.Errorf("healthy response %d %q", w.Code, w.Body.String())
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	called := false
	r.Use(func(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) }, RecoveryWithWriter(nil))
	r.GET("/boom", func(*gin.Context) { called = true })
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))
	if called || rec.Code != http.StatusForbidden {
		t.Errorf("aborted request: handler called %v, status %d", called, rec.Code)
	}
}

func TestRecoveryWithoutWriterStaysSilent(t *testing.T) {
	w := serve(RecoveryWithWriter(nil), func(*gin.Context) { panic("x") }, nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", w.Code)
	}
}

func TestRecoveryDebugLogMasksAuthorization(t *testing.T) {
	gin.SetMode(gin.DebugMode)
	defer gin.SetMode(gin.TestMode)
	var out bytes.Buffer
	r2 := gin.New()
	r2.Use(RecoveryWithWriter(&out))
	r2.GET("/boom", func(*gin.Context) { panic("secret failure") })
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	req.Header.Set("Authorization", "Bearer topsecret")
	req.Header.Set("X-Other", "visible")
	r2.ServeHTTP(httptest.NewRecorder(), req)
	got := out.String()
	if strings.Contains(got, "topsecret") || !strings.Contains(got, "Authorization: *") || !strings.Contains(got, "X-Other: visible") {
		t.Errorf("debug log not masked correctly: %q", got)
	}
}

func TestRecovererLogsFormattedTime(t *testing.T) {
	var out bytes.Buffer
	gin.SetMode(gin.TestMode)
	r := &recoverer{now: func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }}
	r.logger = newTestLogger(&out)
	engine := gin.New()
	engine.Use(r.middleware)
	engine.GET("/boom", func(*gin.Context) { panic("x") })
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/boom", nil))
	if !strings.Contains(out.String(), "[Recovery] 2026/01/02 - 03:04:05 panic recovered:") {
		t.Errorf("log %q", out.String())
	}
}
