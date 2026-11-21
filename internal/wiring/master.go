package wiring

import (
	"context"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device/node"
	"github.com/fsuhrau/automationhub/endpoints"
	"github.com/fsuhrau/automationhub/endpoints/manager"
	endpointmaster "github.com/fsuhrau/automationhub/endpoints/master"
	"github.com/fsuhrau/automationhub/endpoints/web"
	"github.com/fsuhrau/automationhub/hub"
	"github.com/fsuhrau/automationhub/hub/session"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Master is a wired master, ready to run.
type Master struct {
	Server   *hub.Service
	Devices  *hub.DeviceManager
	Nodes    *hub.NodeManager
	Sessions *session.Manager
	API      *API
	DB       *gorm.DB

	logger *logrus.Logger
}

// BuildMaster builds the master of the config exactly once: database, device,
// node and session managers, the services and all endpoints. The master is
// only usable when err is nil.
func BuildMaster(cfg config.Service, opts ...Option) (*Master, error) {
	o, err := resolve(cfg, opts)
	if err != nil {
		return nil, err
	}
	if err := o.openDB(cfg.Database); err != nil {
		return nil, err
	}
	auth, err := buildAuth(o.db, cfg.Auth)
	if err != nil {
		return nil, err
	}
	devices := hub.NewDeviceManager(o.logger, cfg)
	nodes := hub.NewNodeManager(o.logger, o.db)
	m := &Master{
		Devices:  devices,
		Nodes:    nodes,
		Sessions: session.NewManager(o.logger, session.SystemDeps()),
		DB:       o.db,
		logger:   o.logger,
		Server:   hub.NewService(hub.Deps{Logger: o.logger, Devices: devices, Config: cfg, Auth: auth}),
		API:      BuildAPI(Infra{DB: o.db, Devices: devices, Nodes: nodes, Config: cfg, Logger: o.logger}),
	}
	m.Server.RegisterHooks(cfg.Hooks)
	if d, ok := cfg.DeviceManager[node.Manager]; ok && d.Enabled {
		o.logger.Info("adding manager for remote nodes")
		devices.AddHandler(node.NewHandler(o.logger, storage.NewDeviceStore(o.db), nodes, o.db))
	}
	addHandlers(devices, o.handlers)
	return m, addEndpoints(m.Server, m.endpoints(cfg)...)
}

// endpoints lists what the master serves, in the order they are mounted.
func (m *Master) endpoints(cfg config.Service) []endpoints.ServiceEndpoint {
	return []endpoints.ServiceEndpoint{
		m.API.Endpoint,
		manager.New(m.logger, m.Devices, cfg),
		web.New(cfg.Paths),
		endpointmaster.New(cfg, m.Devices, m.Nodes, nil),
		// endpoint for websocket connection
		m.Devices,
	}
}

// Run runs the master until its server fails or ctx is done: the device
// control, the sessions and the node manager run next to the http server.
func (m *Master) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	m.API.Start(ctx)
	return m.Server.RunMaster(ctx, m.Nodes, m.Sessions)
}
