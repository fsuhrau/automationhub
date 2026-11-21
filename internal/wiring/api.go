package wiring

import (
	"context"

	"github.com/fsuhrau/automationhub/endpoints/api"
	hubmcp "github.com/fsuhrau/automationhub/endpoints/mcp"
	"github.com/fsuhrau/automationhub/events"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/sse"
	devicesvc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/devicestore"
)

// API is the REST and MCP side of a master with the parts it shares.
type API struct {
	Endpoint *api.Endpoint
	Broker   *sse.Broker
	Control  *hubcontrol.Service
}

// BuildAPI builds the services of the master, connects the events to the SSE
// broker and returns the endpoint that serves them. Call Start to let the
// device control release idle leases.
func BuildAPI(in Infra) *API {
	rt := newRuntime(in.Devices)
	svc := newServices(in, rt)
	listenForDeviceChanges(in, rt.control)
	api.RegisterSSEBridges(rt.broker)
	return &API{Endpoint: api.New(apiDeps(in, svc, rt)), Broker: rt.broker, Control: rt.control}
}

// Start releases the idle leases of the device control until ctx is done.
func (a *API) Start(ctx context.Context) {
	go a.Control.Run(ctx)
}

func apiDeps(in Infra, svc *services, rt *runtime) api.Deps {
	return api.Deps{
		Logger:    in.Logger.WithField("Service", "Api"),
		NodeURL:   in.Config.NodeUrl,
		Cors:      in.Config.Cors,
		DataDir:   in.Config.Paths.TestData,
		Projects:  svc.projects,
		Access:    svc.access,
		Apps:      svc.apps,
		Tests:     svc.tests,
		Runs:      svc.runs,
		Protocols: svc.protocols,
		Overview:  svc.overview,
		Devices:   svc.devices,
		Control:   rt.control,
		Broker:    rt.broker,
		MCP:       newMCP(in, svc, rt),
	}
}

// newMCP builds the MCP tools on the services of this master.
func newMCP(in Infra, svc *services, rt *runtime) *hubmcp.Handler {
	return hubmcp.NewHandler(hubmcp.Deps{
		Runs:      svc.runs,
		Protocols: svc.runs,
		Lookup:    svc.runs,
		Apps:      svc.apps,
		Tests:     svc.tests,
		Devices:   svc.mcpDevices,
		Pool:      in.Devices,
		Control:   rt.control,
		Tracker:   rt.tracker,
		NodeURL:   in.Config.NodeUrl,
		DataDir:   in.Config.Paths.TestData,
		StoreDir:  in.Config.Paths.AppStore,
	})
}

// listenForDeviceChanges tells the UI to reload the device list when a
// device got or lost its live controller, or was locked or unlocked.
func listenForDeviceChanges(in Infra, control *hubcontrol.Service) {
	notifier := devicesvc.NewListNotifier(devicestore.New(in.DB), events.DeviceListChanged, in.Logger)
	control.OnChange(notifier.DeviceChanged)
	events.DeviceLockChanged.Register(events.HandlerFunc[events.DeviceLockChangedPayload](func(p events.DeviceLockChangedPayload) {
		notifier.DeviceChanged(p.DeviceIdentifier)
	}))
}
