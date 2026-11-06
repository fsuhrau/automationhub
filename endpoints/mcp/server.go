// Package mcp embeds a Model Context Protocol server in the master. It is
// mounted per project at /api/:project_id/mcp, so project access tokens work
// unchanged and every tool is scoped to that project.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

const sessionIdleTimeout = 30 * time.Minute

const instructions = `Automation Hub runs UI tests of games on Android, iOS and other devices.
Orchestration: list_apps -> list_bundles/list_tests -> run_test -> wait_for_test_run -> get_test_run -> get_protocol / get_protocol_logs.
Live control: lock_device first, then screenshot / get_hierarchy / tap / swipe / type_text / press_key, and release_device when done.
Coordinates for tap and swipe are in device screenshot pixels (see the size reported by screenshot).
Elements of a connected Unity app can be addressed by XPath on the scene graph from get_hierarchy.`

// Backend is the part of the REST service the tools share with it.
type Backend interface {
	// ProjectDevices lists the devices visible to the project (with Node).
	ProjectDevices(project *models.Project) ([]models.Device, error)
	// FindDevice returns a device of the project or an error wrapping ErrNotFound.
	FindDevice(project *models.Project, id uint) (*models.Device, error)
	StartTestRun(project *models.Project, testID, binaryID uint, startURL, params string) (*models.TestRun, error)
	CancelTestRun(project *models.Project, runID uint) error
	StoreBundle(application *models.App, filePath string) (*models.AppBinary, error)
}

type Deps struct {
	DB      *gorm.DB
	Backend Backend
	Devices manager.Devices
	Control *hubcontrol.Service
	Runs    *runs.Tracker
	NodeURL string
	Version string
}

// Handler serves MCP for all projects. Each project gets its own streamable
// HTTP handler, so a session id can never cross projects.
type Handler struct {
	deps Deps

	mu       sync.Mutex
	projects map[uint]*mcp.StreamableHTTPHandler
}

func NewHandler(deps Deps) *Handler {
	if deps.Version == "" {
		deps.Version = "dev"
	}
	return &Handler{deps: deps, projects: make(map[uint]*mcp.StreamableHTTPHandler)}
}

type callerKey struct{}

// ServeHTTP handles an MCP request for the project; caller names the
// authenticated user or access token and becomes part of the lease owner.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, project *models.Project, caller string) {
	ctx := context.WithValue(r.Context(), callerKey{}, caller)
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
}

func (h *Handler) newServer(project *models.Project, caller string) *mcp.Server {
	t := &tools{Deps: h.deps, project: project, caller: caller}
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
	t.registerOrchestration(server)
	t.registerControl(server)
	return server
}

// ownerFor names the lease owner: the caller plus the session when there is one.
func (t *tools) ownerFor(ss *mcp.ServerSession) string {
	if ss != nil && ss.ID() != "" {
		return "mcp:" + t.caller + ":" + ss.ID()
	}
	return "mcp:" + t.caller
}

// --- helpers ---

func ptr[T any](v T) *T { return &v }

var (
	readOnly    = &mcp.ToolAnnotations{ReadOnlyHint: true}
	writeSafe   = &mcp.ToolAnnotations{DestructiveHint: ptr(false)}
	destructive = &mcp.ToolAnnotations{DestructiveHint: ptr(true)}
)

// jsonResult returns v as pretty JSON text.
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

func textResult(format string, args ...any) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}}}, nil, nil
}
