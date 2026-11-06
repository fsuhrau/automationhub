package node

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDevicesNotifierCoalescesAndAuthenticates(t *testing.T) {
	var mu sync.Mutex
	var bodies []map[string]string
	var tokens []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/node/devices" || r.Method != http.MethodPost {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		bodies = append(bodies, body)
		tokens = append(tokens, r.Header.Get("X-Auth-Token"))
		mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	token := "node-token"
	n := NewDevicesNotifier(strings.TrimPrefix(srv.URL, "http://"), "node-1", &token)
	for i := 0; i < 5; i++ {
		n.Notify()
	}
	time.Sleep(DevicesNotifierDelay + 300*time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 1 {
		t.Fatalf("expected one coalesced notification, got %d", len(bodies))
	}
	if bodies[0]["node"] != "node-1" || tokens[0] != token {
		t.Fatalf("unexpected notification %v token %q", bodies[0], tokens[0])
	}
}

func TestDevicesNotifierWithoutMasterIsNoop(t *testing.T) {
	n := NewDevicesNotifier("", "node-1", nil)
	if n != nil {
		t.Fatal("expected nil notifier without master")
	}
	n.Notify() // must not panic
}
