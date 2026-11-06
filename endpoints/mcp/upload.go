package mcp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Uploads of builds too large for a tool call: begin_upload returns a single
// use URL, the client PUTs the file to it. The URL lies next to the MCP
// endpoint the client uses (".../mcp/uploads/<id>"). A proxy in front of the
// master (automationhub mcp-proxy) reports that address with X-Forwarded-Host,
// X-Forwarded-Proto and X-Forwarded-Prefix; otherwise the master's own
// address is used.

const (
	uploadTTL = 30 * time.Minute
	// MaxUploadBytes limits one upload.
	MaxUploadBytes int64 = 8 << 30
)

var (
	errUploadNotFound = errors.New("upload not found or expired; call begin_upload again")
	errUploadCaller   = errors.New("upload belongs to another caller")
	errUploadBusy     = errors.New("upload is already in progress")
	errUploadTooLarge = fmt.Errorf("upload is larger than %d bytes", MaxUploadBytes)
)

type pendingUpload struct {
	projectID uint
	appID     uint
	fileName  string
	caller    string
	expires   time.Time
	busy      bool
}

type uploadStore struct {
	mu  sync.Mutex
	m   map[string]*pendingUpload
	now func() time.Time
}

func newUploadStore() *uploadStore {
	return &uploadStore{m: make(map[string]*pendingUpload), now: time.Now}
}

func (s *uploadStore) create(u pendingUpload) (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	id := hex.EncodeToString(b)
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	for k, p := range s.m {
		if !p.busy && now.After(p.expires) {
			delete(s.m, k)
		}
	}
	u.expires = now.Add(uploadTTL)
	s.m[id] = &u
	return id, u.expires, nil
}

// claim marks the upload as in progress for the PUT of caller.
func (s *uploadStore) claim(id string, projectID uint, caller string) (pendingUpload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.m[id]
	if !ok || p.projectID != projectID || s.now().After(p.expires) {
		return pendingUpload{}, errUploadNotFound
	}
	if p.caller != caller {
		return pendingUpload{}, errUploadCaller
	}
	if p.busy {
		return pendingUpload{}, errUploadBusy
	}
	p.busy = true
	return *p, nil
}

// finish ends a claimed upload: a stored one is gone, a failed one can be
// retried until it expires.
func (s *uploadStore) finish(id string, stored bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if stored {
		delete(s.m, id)
	} else if p, ok := s.m[id]; ok {
		p.busy = false
	}
}

// uploadFileName accepts a plain file name, e.g. "Client-1.58.0.apk".
func uploadFileName(name string) (string, error) {
	base := filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	if base == "" || base == "." || base == "/" || base == ".." || strings.HasPrefix(base, ".") {
		return "", fmt.Errorf("invalid file name %q", name)
	}
	return base, nil
}

// uploadURL is the upload address as the client sees the MCP endpoint.
func uploadURL(header http.Header, host string, tls bool, project *models.Project, id string) string {
	scheme := "http"
	if tls {
		scheme = "https"
	}
	prefix := "/api/" + project.Identifier + "/mcp"
	if v := header.Get("X-Forwarded-Proto"); v == "http" || v == "https" {
		scheme = v
	}
	if v := header.Get("X-Forwarded-Host"); v != "" {
		host = v
	}
	if v := header.Get("X-Forwarded-Prefix"); strings.HasPrefix(v, "/") {
		prefix = strings.TrimSuffix(v, "/")
	}
	return fmt.Sprintf("%s://%s%s/uploads/%s", scheme, host, prefix, id)
}

// storeUpload writes the body into a fresh folder and adds it as a bundle.
func (d Deps) storeUpload(application *models.App, fileName string, body io.Reader) (*models.AppBinary, bool, error) {
	if err := os.MkdirAll(apps.AppStoragePath, os.ModePerm); err != nil {
		return nil, false, err
	}
	dir, err := os.MkdirTemp(apps.AppStoragePath, "upload-")
	if err != nil {
		return nil, false, err
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, fileName)
	f, err := os.Create(path)
	if err != nil {
		return nil, false, err
	}
	n, err := io.Copy(f, io.LimitReader(body, MaxUploadBytes+1))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return nil, false, fmt.Errorf("receive file: %w", err)
	}
	if n > MaxUploadBytes {
		return nil, false, errUploadTooLarge
	}
	if n == 0 {
		return nil, false, errors.New("the upload is empty")
	}

	bundle, err := d.Backend.StoreBundle(application, path)
	if bundle != nil && err != nil {
		// an identical build exists already
		return bundle, true, nil
	}
	return bundle, false, err
}

// --- tool ---

type BeginUploadArgs struct {
	AppID    uint   `json:"appId" jsonschema:"app id from list_apps"`
	FileName string `json:"fileName" jsonschema:"file name of the build, e.g. Client-1.58.0.apk"`
}

func (t *tools) beginUpload(ctx context.Context, req *mcp.CallToolRequest, in BeginUploadArgs) (*mcp.CallToolResult, any, error) {
	application, err := t.findApp(in.AppID)
	if err != nil {
		return nil, nil, err
	}
	name, err := uploadFileName(in.FileName)
	if err != nil {
		return nil, nil, err
	}
	id, expires, err := t.uploads.create(pendingUpload{projectID: t.project.ID, appID: application.ID, fileName: name, caller: t.caller})
	if err != nil {
		return nil, nil, err
	}
	var header http.Header
	if req != nil && req.Extra != nil {
		header = req.Extra.Header
	}
	if header == nil {
		header = http.Header{}
	}
	u := uploadURL(header, header.Get(originalHostHeader), header.Get(originalTLSHeader) == "1", t.project, id)
	return jsonResult(map[string]any{
		"uploadUrl": u,
		"method":    "PUT",
		"expiresAt": expires,
		"maxBytes":  MaxUploadBytes,
		"example":   fmt.Sprintf("curl --fail-with-body -T '<path to %s>' '%s'", name, u),
		"note": "Send the raw file as the PUT body, once. Through the AutomationHub host proxy no auth header is needed; " +
			"when connected to the master directly, send the same X-Auth-Token as for MCP. The response is the stored bundle.",
	})
}

// --- HTTP ---

// The master's own address, set by ServeHTTP for URLs without a proxy.
const (
	originalHostHeader = "X-Automationhub-Host"
	originalTLSHeader  = "X-Automationhub-Tls"
)

// ServeUpload receives the PUT of an upload started with begin_upload.
func (h *Handler) ServeUpload(w http.ResponseWriter, r *http.Request, project *models.Project, caller, id string) {
	writeJSON := func(status int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}
	fail := func(status int, err error) {
		writeJSON(status, map[string]string{"message": err.Error()})
	}
	if caller == "" {
		caller = "anonymous"
	}

	if r.ContentLength > MaxUploadBytes {
		fail(http.StatusRequestEntityTooLarge, errUploadTooLarge)
		return
	}
	upload, err := h.uploads.claim(id, project.ID, caller)
	switch {
	case errors.Is(err, errUploadNotFound):
		fail(http.StatusNotFound, err)
		return
	case errors.Is(err, errUploadCaller):
		fail(http.StatusForbidden, err)
		return
	case err != nil:
		fail(http.StatusConflict, err)
		return
	}

	stored := false
	defer func() { h.uploads.finish(id, stored) }()

	var application models.App
	if err := h.deps.DB.First(&application, "id = ? AND project_id = ?", upload.appID, project.ID).Error; err != nil {
		fail(http.StatusNotFound, fmt.Errorf("app %d: %w", upload.appID, err))
		return
	}
	bundle, exists, err := h.deps.storeUpload(&application, upload.fileName, r.Body)
	if errors.Is(err, errUploadTooLarge) {
		fail(http.StatusRequestEntityTooLarge, err)
		return
	}
	if err != nil {
		fail(http.StatusUnprocessableEntity, err)
		return
	}
	stored = true
	if exists {
		writeJSON(http.StatusOK, map[string]any{"exists": true, "bundle": toBundleInfo(bundle)})
		return
	}
	writeJSON(http.StatusCreated, toBundleInfo(bundle))
}
