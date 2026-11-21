// Package api is the REST endpoint of the hub: it mounts the handlers of
// every resource under /api. It builds nothing itself; the composition root
// hands it the services in Deps.
package api

import (
	appsapi "github.com/fsuhrau/automationhub/endpoints/api/apps"
	bundlesapi "github.com/fsuhrau/automationhub/endpoints/api/bundles"
	devicesapi "github.com/fsuhrau/automationhub/endpoints/api/devices"
	overviewapi "github.com/fsuhrau/automationhub/endpoints/api/overview"
	profileapi "github.com/fsuhrau/automationhub/endpoints/api/profile"
	projectsapi "github.com/fsuhrau/automationhub/endpoints/api/projects"
	protocolsapi "github.com/fsuhrau/automationhub/endpoints/api/protocols"
	runsapi "github.com/fsuhrau/automationhub/endpoints/api/runs"
	settingsapi "github.com/fsuhrau/automationhub/endpoints/api/settings"
	testsapi "github.com/fsuhrau/automationhub/endpoints/api/tests"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/gin-gonic/gin"
)

// Endpoint serves the REST API. It implements endpoints.ServiceEndpoint.
type Endpoint struct {
	deps    Deps
	devices *devicesapi.Handler
}

// New returns the endpoint on the services of deps.
func New(deps Deps) *Endpoint {
	return &Endpoint{deps: deps, devices: devicesapi.New(deps.Devices)}
}

// RegisterRoutes mounts the API below the authenticated group.
func (e *Endpoint) RegisterRoutes(_ *gin.Engine, auth *gin.RouterGroup) error {
	d := e.deps
	resolveProject := httpx.ResolveProject(d.Projects)
	api := auth.Group("/api")
	api.Use(httpx.WithLogger(d.Logger))

	registerSSE(api, d.Broker)
	overviewapi.New(d.Overview).RegisterStats(api.Group("/:project_id", resolveProject))
	profileapi.New(d.Access).Register(api.Group("/me"), requireSession)
	settingsapi.New(d.Access).Register(api.Group("/:project_id/settings", requireSession, resolveProject))
	projectsapi.New(d.Projects).Register(api)
	registerDataRoutes(api, d.DataDir)

	e.registerProjectScope(api.Group("/:project_id", resolveProject))
	return nil
}

// registerProjectScope adds the routes of one project, and below them the
// routes of its apps.
func (e *Endpoint) registerProjectScope(projectApi *gin.RouterGroup) {
	d := e.deps
	e.devices.Register(projectApi)
	e.devices.RegisterProtocols(projectApi)
	overviewapi.New(d.Overview).RegisterOverview(projectApi)
	registerDeviceControlRoutes(projectApi, e.deviceControl())
	registerMCPRoutes(projectApi, d.MCP)
	appsapi.New(d.Apps).RegisterProject(projectApi)

	e.registerAppScope(projectApi.Group("/app/:app_id", httpx.ResolveApp(d.Apps)))
}

// registerAppScope adds the routes of one app of the project.
func (e *Endpoint) registerAppScope(appApi *gin.RouterGroup) {
	d := e.deps
	appsapi.New(d.Apps).RegisterApp(appApi)
	bundlesapi.New(d.Apps).Register(appApi)
	testsapi.New(d.Tests).Register(appApi)
	runsapi.New(d.Runs).Register(appApi)
	protocolsapi.New(d.Protocols).RegisterProtocols(appApi)
	protocolsapi.New(d.Protocols).RegisterBaselines(appApi)
}
