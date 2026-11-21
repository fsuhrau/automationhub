package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	sse "github.com/r3labs/sse/v2"
	"github.com/spf13/viper"
)

func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

// execute runs the command line and returns what it printed.
func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()
	resetViper(t)
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs(append([]string{"--config", emptyConfig(t)}, args...))
	err := root.Execute()
	return out.String(), err
}

func emptyConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cli.yaml")
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// fakeHub is a hub API (and event streams) served on httptest.
type fakeHub struct {
	*httptest.Server
	events *sse.Server

	mu       sync.Mutex
	requests []string
	tokens   []string
	failures map[string]int
}

const appPrefix = "/api/demo/app/7"

func newFakeHub(t *testing.T) *fakeHub {
	t.Helper()
	h := &fakeHub{events: sse.New(), failures: map[string]int{}}
	t.Cleanup(h.events.Close)
	mux := http.NewServeMux()
	mux.HandleFunc(appPrefix+"/tests", h.reply(`[{"ID":3,"name":"smoke"},{"ID":4,"name":"login"}]`))
	mux.HandleFunc(appPrefix+"/test/3/run", h.reply(`{"ID":5}`))
	mux.HandleFunc(appPrefix+"/upload", h.reply(`{"ID":11}`))
	mux.HandleFunc(appPrefix+"/bundle/11", h.reply(`{}`))
	mux.HandleFunc("/api/sse", h.events.ServeHTTP)
	h.Server = httptest.NewServer(mux)
	t.Cleanup(h.Close)
	return h
}

// reply answers with body, or with the status registered through failOn.
func (h *fakeHub) reply(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		h.mu.Lock()
		h.requests = append(h.requests, r.Method+" "+r.URL.Path+" "+string(data))
		h.tokens = append(h.tokens, r.Header.Get("X-Auth-Token"))
		status := h.failures[r.URL.Path]
		h.mu.Unlock()
		if status != 0 {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"message":"hub says no"}`))
			return
		}
		_, _ = w.Write([]byte(body))
	}
}

func (h *fakeHub) failOn(path string, status int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.failures[path] = status
}

func (h *fakeHub) seen() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.requests...)
}

// publish makes the event available to every later subscriber.
func (h *fakeHub) publish(stream string, payload any) {
	data, _ := json.Marshal(payload)
	h.events.CreateStream(stream)
	h.events.Publish(stream, &sse.Event{Data: data})
}

type fakeStream struct {
	events map[string][]*sse.Event
	fail   map[string]error
}

func (f fakeStream) Subscribe(ctx context.Context, stream string, ch chan *sse.Event) error {
	if err := f.fail[stream]; err != nil {
		return err
	}
	go func() {
		for _, ev := range f.events[stream] {
			select {
			case ch <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}
