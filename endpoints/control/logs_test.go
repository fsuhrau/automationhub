package control

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"context"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
)

// growingLog is a device whose log grows, like logcat does.
type growingLog struct {
	*testkit.FakeDevice
	mu    sync.Mutex
	lines []string
}

func (d *growingLog) add(line string) { d.mu.Lock(); d.lines = append(d.lines, line); d.mu.Unlock() }

func (d *growingLog) ReadLogs(n int) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return slices.Clone(d.lines[max(0, len(d.lines)-n):]), nil
}

func (d *growingLog) ReadLogsSince(string, int) ([]string, error) { return d.ReadLogs(1) }

func TestLogs(t *testing.T) {
	e := setup(t)
	e.full.LogLines = []string{"a", "b", "c"}
	rec := e.call(t, "", http.MethodGet, "1/logs?lines=2", nil)
	if got := testkit.Decode[map[string][]string](t, rec)["lines"]; !slices.Equal(got, []string{"b", "c"}) {
		t.Fatalf("lines %q", got)
	}
}

func TestLogsWithoutLinesIsAnEmptyArray(t *testing.T) {
	e := setup(t)
	rec := e.call(t, "", http.MethodGet, "1/logs", nil)
	if !strings.Contains(rec.Body.String(), `"lines":[]`) {
		t.Fatalf("body %s", rec.Body)
	}
}

func TestLogsReadFailure(t *testing.T) {
	e := setup(t)
	e.full.FailWith("ReadLogs", errors.New("boom"))
	if rec := e.call(t, "", http.MethodGet, "1/logs", nil); rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestLogLines(t *testing.T) {
	tests := map[string]int{"": 200, "10": 10, "0": 200, "-1": 200, "5001": 200, "5000": 5000, "abc": 200}
	for q, want := range tests {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/?lines="+q, nil)
		if got := logLines(c); got != want {
			t.Errorf("lines=%q: got %d, want %d", q, got, want)
		}
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if got := logLines(c); got != 200 {
		t.Errorf("no lines parameter: got %d", got)
	}
}

func TestLogStream(t *testing.T) {
	e := setup(t)
	dev := &growingLog{FakeDevice: testkit.NewFakeDevice("log-3"), lines: []string{
		"11-09 12:00:00.100 I/A( 1): one",
		"11-09 12:00:00.200 I/A( 1): two",
	}}
	e.devices["3"] = dev
	srv := httptest.NewServer(e.router)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/p/device/3/logs/stream?lines=10", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type %q", ct)
	}

	scanner := bufio.NewScanner(resp.Body)
	readLines := func() []string {
		event := ""
		for scanner.Scan() {
			line := scanner.Text()
			if v, ok := strings.CutPrefix(line, "event: "); ok {
				event = v
			} else if v, ok := strings.CutPrefix(line, "data: "); ok && event == "lines" {
				var lines []string
				if err := json.Unmarshal([]byte(v), &lines); err != nil {
					t.Fatal(err)
				}
				return lines
			}
		}
		t.Fatalf("stream ended: %v", scanner.Err())
		return nil
	}

	if got := readLines(); len(got) != 2 {
		t.Fatalf("backlog %q", got)
	}
	dev.add("11-09 12:00:01.000 E/B( 2): three")
	if got := readLines(); !slices.Equal(got, []string{"11-09 12:00:01.000 E/B( 2): three"}) {
		t.Fatalf("follow %q", got)
	}
}

func TestLogStreamNeedsALogReader(t *testing.T) {
	e := setup(t)
	if rec := e.call(t, "", http.MethodGet, "2/logs/stream", nil); rec.Code != http.StatusNotImplemented {
		t.Fatalf("plain device: expected 501, got %d", rec.Code)
	}
}

func TestEventSenderSaysEachFailureOnce(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	s := &eventSender{c: c}
	for _, err := range []error{errors.New("down"), errors.New("down"), errors.New("other")} {
		if e := s.failure(err); e != nil {
			t.Fatal(e)
		}
	}
	if got := strings.Count(rec.Body.String(), "event: failure"); got != 2 {
		t.Fatalf("failures sent %d times: %q", got, rec.Body)
	}
	if err := s.lines([]string{"x"}); err != nil || s.lastFailure != "" {
		t.Fatalf("lines must reset the failure: %v %q", err, s.lastFailure)
	}
	if !strings.Contains(rec.Body.String(), "event: lines\ndata: [\"x\"]\n\n") {
		t.Fatalf("body %q", rec.Body)
	}
}

type errWriter struct{ gin.ResponseWriter }

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestEventSenderReportsWriteErrors(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Writer = errWriter{c.Writer}
	if err := (&eventSender{c: c}).lines([]string{"x"}); err == nil {
		t.Fatal("write error swallowed")
	}
}

func TestLogStreamEndsWithTheClient(t *testing.T) {
	e := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/p/device/1/logs/stream", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Header().Get("X-Accel-Buffering") != "no" {
		t.Fatalf("unexpected response %d %v", rec.Code, rec.Header())
	}
}

func TestLogStreamStopsWhenTheClientWriteFails(t *testing.T) {
	dev := testkit.NewFakeDevice("log-1")
	dev.LogLines = []string{"11-09 12:00:00.100 I/A( 1): one"}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Writer = errWriter{c.Writer}
	done := make(chan struct{})
	go func() { (&Handler{svc: hubcontrol.New(nil)}).logStream(c, dev); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("log stream still running after a write error")
	}
}
