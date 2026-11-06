package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

// uploadBackend records the files handed to StoreBundle.
type uploadBackend struct {
	fakeBackend
	mu     sync.Mutex
	stored map[string]string // file name -> content
}

func (b *uploadBackend) StoreBundle(application *models.App, filePath string) (*models.AppBinary, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.stored[filepath.Base(filePath)] = string(data)
	b.mu.Unlock()
	bundle := &models.AppBinary{Name: "Client", Version: "1.58.0", AppID: application.ID}
	bundle.ID = 42
	return bundle, nil
}

type uploadEnv struct {
	srv     *httptest.Server
	handler *Handler
	backend *uploadBackend
	project *models.Project
}

func uploadSetup(t *testing.T) *uploadEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Project{}, &models.App{}); err != nil {
		t.Fatal(err)
	}
	project := &models.Project{Identifier: "p1", Name: "P1"}
	db.Create(project)
	other := &models.Project{Identifier: "p2", Name: "P2"}
	db.Create(other)
	db.Create(&models.App{Name: "Android", ProjectID: project.ID})
	db.Create(&models.App{Name: "Foreign", ProjectID: other.ID})

	old := apps.AppStoragePath
	apps.AppStoragePath = t.TempDir()
	t.Cleanup(func() { apps.AppStoragePath = old })

	backend := &uploadBackend{stored: map[string]string{}}
	h := NewHandler(Deps{DB: db, Backend: backend, Control: hubcontrol.New(nil), Runs: runs.NewTracker(nil)})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		caller := r.Header.Get("X-Caller")
		if id, ok := strings.CutPrefix(r.URL.Path, "/api/p1/mcp/uploads/"); ok {
			h.ServeUpload(w, r, project, caller, id)
			return
		}
		h.ServeHTTP(w, r, project, caller)
	}))
	t.Cleanup(srv.Close)
	return &uploadEnv{srv: srv, handler: h, backend: backend, project: project}
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
	defer resp.Body.Close()
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
	if resp.StatusCode != http.StatusCreated || !strings.Contains(body, `"id":42`) {
		t.Fatalf("upload: %d %s", resp.StatusCode, body)
	}
	if got := env.backend.stored["Client-1.58.0.apk"]; got != "APK-BYTES" {
		t.Fatalf("stored %v", env.backend.stored)
	}
	// single use; the temporary folder is gone
	if resp, _ := put(t, r.UploadURL, "alice", "again"); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("second upload: %d", resp.StatusCode)
	}
	entries, _ := os.ReadDir(apps.AppStoragePath)
	if len(entries) != 0 {
		t.Fatalf("leftovers in the store: %v", entries)
	}
}

func TestUploadRules(t *testing.T) {
	env := uploadSetup(t)

	cs := connectWith(t, env.srv, headerTransport{caller: "alice"})
	if res := call(t, cs, "begin_upload", map[string]any{"appId": 2, "fileName": "a.apk"}); !res.IsError {
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
	env.handler.ServeUpload(w, req, env.project, "alice", id)
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
