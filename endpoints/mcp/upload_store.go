package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

// create stores the upload under a new random id and returns it with the
// time it expires.
func (s *uploadStore) create(u pendingUpload) (string, time.Time) {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // never fails
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
	return id, u.expires
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
