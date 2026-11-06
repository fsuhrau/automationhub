package mcpproxy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type seen struct {
	path, token, cookie, authorization         string
	method, fwdHost, fwdProto, fwdPrefix, body string
}

func setup(t *testing.T, tokenErr error) (*httptest.Server, chan seen) {
	t.Helper()
	requests := make(chan seen, 10)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests <- seen{r.URL.Path, r.Header.Get("X-Auth-Token"), r.Header.Get("Cookie"), r.Header.Get("Authorization"),
			r.Method, r.Header.Get("X-Forwarded-Host"), r.Header.Get("X-Forwarded-Proto"), r.Header.Get("X-Forwarded-Prefix"), string(body)}
		w.Header().Set("Set-Cookie", "session=secret")
		if r.Method == http.MethodGet {
			// an MCP event stream: the first event must arrive before the end
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "data: first\n\n")
			w.(http.Flusher).Flush()
			time.Sleep(500 * time.Millisecond)
			fmt.Fprint(w, "data: second\n\n")
			return
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(upstream.Close)
	u, _ := url.Parse(upstream.URL)
	source := TokenFunc(func() (string, error) {
		if tokenErr != nil {
			return "", tokenErr
		}
		return "ahp_host-token", nil
	})
	proxy := httptest.NewServer(Handler(Config{Upstream: u, Project: "game", Token: source, Key: "k1", Logf: t.Logf}))
	t.Cleanup(proxy.Close)
	return proxy, requests
}

func send(t *testing.T, method, url, key string, headers ...string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(method, url, strings.NewReader(`{}`))
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestForwardsMCPWithHostToken(t *testing.T) {
	proxy, requests := setup(t, nil)

	resp := send(t, "POST", proxy.URL+"/mcp", "k1", "Cookie", "session=container", "X-Auth-Token", "ahp_from-container")
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Set-Cookie") != "" {
		t.Fatalf("status %d, set-cookie %q", resp.StatusCode, resp.Header.Get("Set-Cookie"))
	}
	got := <-requests
	if got.path != "/api/game/mcp" || got.token != "ahp_host-token" || got.cookie != "" || got.authorization != "" {
		t.Fatalf("upstream saw %+v", got)
	}

	send(t, "DELETE", proxy.URL+"/other-project/mcp", "k1")
	if got := <-requests; got.path != "/api/other-project/mcp" {
		t.Fatalf("project path: %+v", got)
	}
}

func TestOnlyMCPIsForwarded(t *testing.T) {
	proxy, requests := setup(t, nil)
	for _, path := range []string{"/api/game/devices", "/game/devices", "/a/b/mcp", "/%2e%2e/mcp", "/", "/mcp/x"} {
		if resp := send(t, "POST", proxy.URL+path, "k1"); resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", path, resp.StatusCode)
		}
	}
	if resp := send(t, "PUT", proxy.URL+"/mcp", "k1"); resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("PUT: expected 405, got %d", resp.StatusCode)
	}
	select {
	case r := <-requests:
		t.Fatalf("forwarded %+v", r)
	default:
	}
}

func TestProxyKey(t *testing.T) {
	proxy, requests := setup(t, nil)
	for _, key := range []string{"", "wrong"} {
		if resp := send(t, "POST", proxy.URL+"/mcp", key); resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("key %q: expected 401, got %d", key, resp.StatusCode)
		}
	}
	select {
	case r := <-requests:
		t.Fatalf("forwarded without key: %+v", r)
	default:
	}
}

func TestMissingHostToken(t *testing.T) {
	proxy, requests := setup(t, fmt.Errorf("keychain: %w", ErrNoToken))
	resp := send(t, "POST", proxy.URL+"/mcp", "k1")
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusBadGateway || !strings.Contains(string(body), "no AutomationHub token") {
		t.Fatalf("%d %s", resp.StatusCode, body)
	}
	if len(requests) != 0 {
		t.Fatal("forwarded without a token")
	}
}

func TestStreamsEvents(t *testing.T) {
	proxy, _ := setup(t, nil)
	start := time.Now()
	resp := send(t, "GET", proxy.URL+"/mcp", "k1")
	line, err := bufio.NewReader(resp.Body).ReadString('\n')
	if err != nil || line != "data: first\n" {
		t.Fatalf("first event %q %v", line, err)
	}
	if time.Since(start) > 400*time.Millisecond {
		t.Fatal("the event was buffered until the stream ended")
	}
}

func TestTokenSources(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "token")
	if _, err := FileToken(file).Token(); !errors.Is(err, ErrNoToken) {
		t.Fatalf("missing file: %v", err)
	}
	os.WriteFile(file, []byte("ahp_file\n"), 0o644)
	if _, err := FileToken(file).Token(); err == nil || errors.Is(err, ErrNoToken) || !strings.Contains(err.Error(), "chmod 600") {
		t.Fatalf("readable file must be refused: %v", err)
	}
	// a refused file stops the search instead of being skipped silently
	if _, err := FirstToken(FileToken(file), EnvToken("MCPPROXY_TEST_UNSET")).Token(); err == nil || errors.Is(err, ErrNoToken) {
		t.Fatalf("misconfigured file skipped: %v", err)
	}
	os.Chmod(file, 0o600)
	t.Setenv("MCPPROXY_TEST_TOKEN", "ahp_env")
	if v, err := FirstToken(EnvToken("MCPPROXY_TEST_UNSET"), FileToken(file)).Token(); err != nil || v != "ahp_file" {
		t.Fatalf("file: %q %v", v, err)
	}
	if v, _ := FirstToken(EnvToken("MCPPROXY_TEST_TOKEN"), FileToken(file)).Token(); v != "ahp_env" {
		t.Fatalf("env first: %q", v)
	}
	_, err := FirstToken(EnvToken("MCPPROXY_TEST_UNSET"), FileToken(filepath.Join(dir, "none"))).Token()
	if !errors.Is(err, ErrNoToken) || !strings.Contains(err.Error(), "MCPPROXY_TEST_UNSET") {
		t.Fatalf("nothing found: %v", err)
	}
}

func TestCachedToken(t *testing.T) {
	var calls atomic.Int32
	source := CachedToken(TokenFunc(func() (string, error) {
		return fmt.Sprintf("t%d", calls.Add(1)), nil
	}), 50*time.Millisecond)
	a, _ := source.Token()
	b, _ := source.Token()
	time.Sleep(60 * time.Millisecond)
	c, _ := source.Token()
	if a != "t1" || b != "t1" || c != "t2" {
		t.Fatalf("got %s %s %s", a, b, c)
	}
}

func TestLoadOrCreateKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "proxy-key")
	k1, err := LoadOrCreateKey(path)
	if err != nil || len(k1) != 64 {
		t.Fatalf("%q %v", k1, err)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o", info.Mode().Perm())
	}
	if k2, _ := LoadOrCreateKey(path); k2 != k1 {
		t.Fatal("key changed")
	}
}

const uploadID = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestForwardedAddressForUploadURLs(t *testing.T) {
	proxy, requests := setup(t, nil)
	send(t, "POST", proxy.URL+"/game/mcp", "k1", "X-Forwarded-Host", "evil.example", "X-Forwarded-Prefix", "/evil")
	got := <-requests
	host := strings.TrimPrefix(proxy.URL, "http://")
	if got.fwdHost != host || got.fwdProto != "http" || got.fwdPrefix != "/game/mcp" {
		t.Fatalf("forwarded %+v, want host %s", got, host)
	}
	send(t, "POST", proxy.URL+"/mcp", "k1")
	if got := <-requests; got.fwdPrefix != "/mcp" {
		t.Fatalf("prefix of /mcp: %q", got.fwdPrefix)
	}
}

func TestForwardsUploads(t *testing.T) {
	proxy, requests := setup(t, nil)

	// the upload id is the credential: no proxy key needed
	resp := send(t, "PUT", proxy.URL+"/game/mcp/uploads/"+uploadID, "", "Cookie", "c=1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload: %d", resp.StatusCode)
	}
	got := <-requests
	if got.method != "PUT" || got.path != "/api/game/mcp/uploads/"+uploadID || got.token != "ahp_host-token" || got.cookie != "" || got.body != "{}" {
		t.Fatalf("upstream saw %+v", got)
	}
	send(t, "PUT", proxy.URL+"/mcp/uploads/"+uploadID, "")
	if got := <-requests; got.path != "/api/game/mcp/uploads/"+uploadID {
		t.Fatalf("default project upload: %+v", got)
	}

	for _, c := range []struct {
		method, path string
		status       int
	}{
		{"POST", "/game/mcp/uploads/" + uploadID, http.StatusMethodNotAllowed},
		{"GET", "/game/mcp/uploads/" + uploadID, http.StatusMethodNotAllowed},
		{"PUT", "/game/mcp", http.StatusMethodNotAllowed},
		{"PUT", "/game/mcp/uploads/short", http.StatusNotFound},
		{"PUT", "/game/mcp/uploads/" + uploadID + "/x", http.StatusNotFound},
		{"PUT", "/game/devices/uploads/" + uploadID, http.StatusNotFound},
	} {
		if resp := send(t, c.method, proxy.URL+c.path, ""); resp.StatusCode != c.status {
			t.Errorf("%s %s: expected %d, got %d", c.method, c.path, c.status, resp.StatusCode)
		}
	}
	select {
	case r := <-requests:
		t.Fatalf("forwarded %+v", r)
	default:
	}
}
