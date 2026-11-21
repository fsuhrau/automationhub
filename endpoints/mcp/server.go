// Package mcp embeds a Model Context Protocol server in the master. It is
// mounted per project at /api/:project_id/mcp, so project access tokens work
// unchanged and every tool is scoped to that project.
package mcp

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const sessionIdleTimeout = 30 * time.Minute

const instructions = `Automation Hub runs UI tests of games on Android, iOS and other devices.
Orchestration: list_apps -> list_bundles/list_tests -> run_test -> wait_for_test_run -> get_test_run -> get_protocol / get_protocol_logs.
Live control: lock_device first, then screenshot / get_hierarchy / tap / swipe / type_text / press_key, and release_device when done.
Coordinates for tap and swipe are in device screenshot pixels (see the size reported by screenshot).
Elements of a connected Unity app can be addressed by XPath on the scene graph from get_hierarchy.
Builds: upload_bundle fetches a build from an http(s) URL; for a local file call begin_upload and PUT the file to the returned uploadUrl (e.g. curl -T).`

// Deps are what the tools call: the services of the master, the pool of live
// devices and the live control. The tools hold no database.
type Deps struct {
	Runs      RunService
	Protocols ProtocolService
	Lookup    DeviceLookup
	Apps      AppService
	Tests     TestService
	Devices   DeviceService
	Pool      Pool
	Control   Control
	Tracker   *runs.Tracker
	NodeURL   string
	Version   string
	// DataDir holds the screenshot files of protocols and StoreDir the bundle
	// files; they default to the folders of the storage.
	DataDir  string
	StoreDir string
	// MaxUpload limits one upload in bytes; it defaults to MaxUploadBytes.
	MaxUpload int64
	// ProgressEvery is how often wait_for_test_run reports progress; it
	// defaults to 10 seconds.
	ProgressEvery time.Duration
}

// Handler serves MCP for all projects. Each project gets its own streamable
// HTTP handler, so a session id can never cross projects.
type Handler struct {
	deps Deps

	mu       sync.Mutex
	projects map[uint]*mcp.StreamableHTTPHandler
	uploads  *uploadStore
}

func NewHandler(deps Deps) *Handler {
	if deps.Version == "" {
		deps.Version = "dev"
	}
	if deps.DataDir == "" {
		deps.DataDir = apps.TestDataPath
	}
	if deps.StoreDir == "" {
		deps.StoreDir = apps.AppStoragePath
	}
	if deps.MaxUpload <= 0 {
		deps.MaxUpload = MaxUploadBytes
	}
	if deps.ProgressEvery <= 0 {
		deps.ProgressEvery = 10 * time.Second
	}
	return &Handler{deps: deps, projects: make(map[uint]*mcp.StreamableHTTPHandler), uploads: newUploadStore()}
}

type callerKey struct{}

// ServeHTTP handles an MCP request for the project; caller names the
// authenticated user or access token and becomes part of the lease owner.
//
//nolint:revive // called by the MCP route of endpoints/api with these arguments
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, project *models.Project, caller string) {
	ctx := context.WithValue(r.Context(), callerKey{}, caller)
	// the master's own address, for upload URLs when no proxy reports one
	r.Header.Set(originalHostHeader, r.Host)
	if r.TLS != nil {
		r.Header.Set(originalTLSHeader, "1")
	} else {
		r.Header.Del(originalTLSHeader)
	}
	h.forProject(project).ServeHTTP(w, r.WithContext(ctx))
}

func (h *Handler) forProject(project *models.Project) *mcp.StreamableHTTPHandler {
	h.mu.Lock()
	defer h.mu.Unlock()
	if hh, ok := h.projects[project.ID]; ok {
		return hh
	}
	p := *project
	hh := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		caller, _ := r.Context().Value(callerKey{}).(string)
		if caller == "" {
			caller = "anonymous"
		}
		return h.newServer(&p, caller)
	}, &mcp.StreamableHTTPOptions{SessionTimeout: sessionIdleTimeout})
	h.projects[project.ID] = hh
	return hh
}

// tools holds everything a tool call needs; one instance per MCP server.
type tools struct {
	Deps
	project *models.Project
	caller  string
	uploads *uploadStore
}

func (h *Handler) newServer(project *models.Project, caller string) *mcp.Server {
	t := &tools{Deps: h.deps, project: project, caller: caller, uploads: h.uploads}
	server := mcp.NewServer(&mcp.Implementation{Name: "automationhub", Version: h.deps.Version}, &mcp.ServerOptions{
		Instructions: instructions,
		InitializedHandler: func(ctx context.Context, req *mcp.InitializedRequest) {
			ss := req.Session
			if ss == nil || ss.ID() == "" {
				return
			}
			owner := t.ownerFor(ss)
			go func() {
				_ = ss.Wait()
				// a closed or timed out session gives up its devices
				t.Control.ReleaseOwner(owner)
			}()
		},
	})
	for _, register := range []func(*mcp.Server){
		t.registerDevices, t.registerApps, t.registerRuns, t.registerProtocols, t.registerUploads, t.registerControl,
	} {
		register(server)
	}
	return server
}

// ownerFor names the lease owner: the caller plus the session when there is one.
func (t *tools) ownerFor(ss *mcp.ServerSession) string {
	if ss != nil && ss.ID() != "" {
		return "mcp:" + t.caller + ":" + ss.ID()
	}
	return "mcp:" + t.caller
}
