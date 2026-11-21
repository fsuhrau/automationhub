package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/fsuhrau/automationhub/storage/models"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type uploadEnv struct {
	srv     *httptest.Server
	handler *Handler
	world   *goldenWorld
}

func uploadSetup(t *testing.T) *uploadEnv {
	t.Helper()
	w := newGoldenWorld(t)
	h := w.handler()
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		caller := r.Header.Get("X-Caller")
		if id, ok := strings.CutPrefix(r.URL.Path, "/api/p1/mcp/uploads/"); ok {
			h.ServeUpload(rw, r, w.project, caller, id)
			return
		}
		h.ServeHTTP(rw, r, w.project, caller)
	}))
	t.Cleanup(srv.Close)
	return &uploadEnv{srv: srv, handler: h, world: w}
}

// headerTransport adds headers to every request, e.g. the X-Forwarded-*
// headers of the host proxy.
type headerTransport struct {
	caller  string
	headers map[string]string
}

func (h headerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("X-Caller", h.caller)
	for k, v := range h.headers {
		r.Header.Set(k, v)
	}
	return http.DefaultTransport.RoundTrip(r)
}

func connectWith(t *testing.T, srv *httptest.Server, rt http.RoundTripper) *mcp.ClientSession {
	t.Helper()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	cs, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint:   srv.URL,
		HTTPClient: &http.Client{Transport: rt},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

type beginResult struct {
	UploadURL string `json:"uploadUrl"`
	Method    string `json:"method"`
}

func begin(t *testing.T, env *uploadEnv, caller string, headers map[string]string, args map[string]any) beginResult {
	t.Helper()
	cs := connectWith(t, env.srv, headerTransport{caller, headers})
	res := call(t, cs, "begin_upload", args)
	if res.IsError {
		t.Fatalf("begin_upload: %s", text(res))
	}
	var r beginResult
	if err := json.Unmarshal([]byte(text(res)), &r); err != nil {
		t.Fatalf("result %s: %v", text(res), err)
	}
	return r
}

func put(t *testing.T, url, caller, body string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPut, url, strings.NewReader(body))
	req.Header.Set("X-Caller", caller)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, _ := io.ReadAll(resp.Body)
	return resp, string(data)
}

var uploadIDInURL = regexp.MustCompile(`/uploads/[a-f0-9]{64}$`)

func TestUploadURLFollowsTheProxy(t *testing.T) {
	env := uploadSetup(t)
	r := begin(t, env, "alice", map[string]string{
		"X-Forwarded-Host":   "automationhub-mcp:18092",
		"X-Forwarded-Proto":  "http",
		"X-Forwarded-Prefix": "/p1/mcp",
	}, map[string]any{"appId": 1, "fileName": "Client-1.58.0.apk"})
	if !strings.HasPrefix(r.UploadURL, "http://automationhub-mcp:18092/p1/mcp/uploads/") || !uploadIDInURL.MatchString(r.UploadURL) || r.Method != "PUT" {
		t.Fatalf("upload url %+v", r)
	}

	// without a proxy the master's own address is used
	r = begin(t, env, "alice", nil, map[string]any{"appId": 1, "fileName": "Client.apk"})
	want := env.srv.URL + "/api/p1/mcp/uploads/"
	if !strings.HasPrefix(r.UploadURL, want) || !uploadIDInURL.MatchString(r.UploadURL) {
		t.Fatalf("direct upload url %q, want prefix %q", r.UploadURL, want)
	}
}

func TestUploadStoresTheBundleOnce(t *testing.T) {
	env := uploadSetup(t)
	r := begin(t, env, "alice", nil, map[string]any{"appId": 1, "fileName": "dir/../Client-1.58.0.apk"})

	if resp, _ := put(t, r.UploadURL, "mallory", "evil"); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("other caller: %d", resp.StatusCode)
	}
	resp, body := put(t, r.UploadURL, "alice", "APK-BYTES")
	if resp.StatusCode != http.StatusCreated || !strings.Contains(body, `"id":4,`) {
		t.Fatalf("upload: %d %s", resp.StatusCode, body)
	}
	entries, _ := os.ReadDir(env.world.storeDir)
	if len(entries) != 1 || !strings.HasPrefix(entries[0].Name(), "Client-1.58.0_") {
		t.Fatalf("stored %v, want the bundle file", entries)
	}
	if got, _ := os.ReadFile(filepath.Join(env.world.storeDir, entries[0].Name())); string(got) != "APK-BYTES" {
		t.Fatalf("stored content %q", got)
	}
	// single use; the temporary folder is gone
	if resp, _ := put(t, r.UploadURL, "alice", "again"); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("second upload: %d", resp.StatusCode)
	}
	if left, _ := os.ReadDir(env.world.storeDir); len(left) != 1 {
		t.Fatalf("leftovers in the store: %v", left)
	}
}

func TestUploadRules(t *testing.T) {
	env := uploadSetup(t)

	cs := connectWith(t, env.srv, headerTransport{caller: "alice"})
	if res := call(t, cs, "begin_upload", map[string]any{"appId": 3, "fileName": "a.apk"}); !res.IsError {
		t.Fatal("app of another project accepted")
	}
	if res := call(t, cs, "begin_upload", map[string]any{"appId": 1, "fileName": ".."}); !res.IsError {
		t.Fatal("invalid file name accepted")
	}

	if resp, _ := put(t, env.srv.URL+"/api/p1/mcp/uploads/"+strings.Repeat("0", 64), "alice", "x"); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown id: %d", resp.StatusCode)
	}

	r := begin(t, env, "alice", nil, map[string]any{"appId": 1, "fileName": "a.apk"})
	if resp, _ := put(t, r.UploadURL, "alice", ""); resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("empty upload: %d", resp.StatusCode)
	}

	// a failed upload can be retried, an expired one not
	env.handler.uploads.now = func() time.Time { return time.Now().Add(uploadTTL + time.Minute) }
	if resp, _ := put(t, r.UploadURL, "alice", "data"); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expired upload: %d", resp.StatusCode)
	}
}

func TestUploadTooLarge(t *testing.T) {
	env := uploadSetup(t)
	r := begin(t, env, "alice", nil, map[string]any{"appId": 1, "fileName": "a.apk"})
	id := r.UploadURL[strings.LastIndex(r.UploadURL, "/")+1:]

	req := httptest.NewRequest(http.MethodPut, "/api/p1/mcp/uploads/"+id, strings.NewReader("x"))
	req.ContentLength = MaxUploadBytes + 1
	w := httptest.NewRecorder()
	env.handler.ServeUpload(w, req, env.world.project, "alice", id)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", w.Code)
	}
}

func TestUploadFileName(t *testing.T) {
	for in, want := range map[string]string{
		"Client.apk":         "Client.apk",
		"/tmp/x/Client.apk":  "Client.apk",
		`C:\builds\Game.ipa`: "Game.ipa",
		"..":                 "",
		".hidden.apk":        "",
		"":                   "",
	} {
		got, err := uploadFileName(in)
		if (want == "") != (err != nil) || got != want {
			t.Errorf("uploadFileName(%q) = %q, %v", in, got, err)
		}
	}
}

func TestUploadOfABuildThatExistsAnswersWithIt(t *testing.T) {
	env := uploadSetup(t)
	first := begin(t, env, "alice", nil, map[string]any{"appId": 1, "fileName": "a.apk"})
	put(t, first.UploadURL, "alice", "SAME")
	second := begin(t, env, "alice", nil, map[string]any{"appId": 1, "fileName": "b.apk"})

	resp, body := put(t, second.UploadURL, "alice", "SAME")

	if resp.StatusCode != http.StatusOK || !strings.Contains(body, `"exists":true`) {
		t.Fatalf("second upload: %d %s", resp.StatusCode, body)
	}
}

func TestUploadForAnAppThatIsGoneIsNotFound(t *testing.T) {
	env := uploadSetup(t)
	r := begin(t, env, "alice", nil, map[string]any{"appId": 1, "fileName": "a.apk"})
	env.world.db.Delete(&models.App{}, 1)

	resp, _ := put(t, r.UploadURL, "alice", "data")

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", resp.StatusCode)
	}
}

func TestUploadLargerThanTheLimitIsRefused(t *testing.T) {
	w := newGoldenWorld(t)
	h := NewHandler(Deps{Runs: w.runs, Protocols: w.runs, Lookup: w.runs, Apps: w.apps, Tests: w.tests, Devices: w.devices,
		Pool: w.pool, Control: w.control, Tracker: w.tracker, StoreDir: w.storeDir, MaxUpload: 4})
	env := &uploadEnv{handler: h, world: w}
	env.srv = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if id, ok := strings.CutPrefix(r.URL.Path, "/api/p1/mcp/uploads/"); ok {
			h.ServeUpload(rw, r, w.project, r.Header.Get("X-Caller"), id)
			return
		}
		h.ServeHTTP(rw, r, w.project, r.Header.Get("X-Caller"))
	}))
	t.Cleanup(env.srv.Close)
	r := begin(t, env, "alice", nil, map[string]any{"appId": 1, "fileName": "a.apk"})

	resp, _ := put(t, r.UploadURL, "alice", "12345")

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status %d", resp.StatusCode)
	}
}

func TestDownloadFailsForAnAddressThatIsNotARequest(t *testing.T) {
	_, err := download(context.Background(), &url.URL{Scheme: "http", Host: "a b"})

	if err == nil {
		t.Fatal("download started")
	}
}

func TestUploadBundleReportsAnUnreachableServerAndAnEmptyFile(t *testing.T) {
	srv, _ := setup(t)
	cs := connect(t, srv, "alice")
	gone := httptest.NewServer(http.NotFoundHandler())
	gone.Close()
	empty := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}))
	defer empty.Close()

	unreachable := call(t, cs, "upload_bundle", map[string]any{"appId": 1, "url": gone.URL + "/a.apk"})
	nothing := call(t, cs, "upload_bundle", map[string]any{"appId": 1, "url": empty.URL + "/a.apk"})

	if !unreachable.IsError || !nothing.IsError || !strings.Contains(text(nothing), "the upload is empty") {
		t.Fatalf("unreachable %q, empty %q", text(unreachable), text(nothing))
	}
}

// putStreamed sends the body without a content length, so the master only
// learns its size by reading it.
func putStreamed(t *testing.T, url, body string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPut, url, io.NopCloser(strings.NewReader(body)))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}

func TestUploadLargerThanTheLimitIsRefusedWhileReceiving(t *testing.T) {
	w := newGoldenWorld(t)
	h := NewHandler(Deps{Apps: w.apps, Runs: w.runs, StoreDir: w.storeDir, MaxUpload: 4})
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		h.ServeUpload(rw, r, w.project, "alice", r.URL.Path[len("/"):])
	}))
	defer srv.Close()
	id, _ := h.uploads.create(pendingUpload{projectID: w.project.ID, appID: 1, fileName: "a.apk", caller: "alice"})

	if got := putStreamed(t, srv.URL+"/"+id, "12345"); got != http.StatusRequestEntityTooLarge {
		t.Fatalf("status %d", got)
	}
}

func TestUploadWithoutACallerIsTheAnonymousCallers(t *testing.T) {
	env := uploadSetup(t)
	r := begin(t, env, "", nil, map[string]any{"appId": 1, "fileName": "a.apk"})

	resp, body := put(t, r.UploadURL, "", "DATA")

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("upload: %d %s", resp.StatusCode, body)
	}
}

func TestUploadStoreForgetsExpiredUploadsAndRefusesAClaimInProgress(t *testing.T) {
	now := time.Now()
	s := newUploadStore()
	s.now = func() time.Time { return now }
	old, _ := s.create(pendingUpload{projectID: 1, caller: "alice"})
	now = now.Add(uploadTTL + time.Minute)

	fresh, _ := s.create(pendingUpload{projectID: 1, caller: "alice"})

	if _, err := s.claim(old, 1, "alice"); !errors.Is(err, errUploadNotFound) {
		t.Errorf("expired upload: %v", err)
	}
	if _, err := s.claim(fresh, 1, "alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.claim(fresh, 1, "alice"); !errors.Is(err, errUploadBusy) {
		t.Errorf("second claim: %v", err)
	}
}
