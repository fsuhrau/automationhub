package wiring

import (
	"context"
	"errors"

	"github.com/fsuhrau/automationhub/config"
	devicenode "github.com/fsuhrau/automationhub/device/node"
	"github.com/fsuhrau/automationhub/endpoints/manager"
	endpointnode "github.com/fsuhrau/automationhub/endpoints/node"
	"github.com/fsuhrau/automationhub/hub"
	hubnode "github.com/fsuhrau/automationhub/hub/node"
)

// ErrNoMasterURL is returned when a node has no master to connect to.
var ErrNoMasterURL = errors.New("slave need a master url to connect to")

// Node is a wired node, ready to run.
type Node struct {
	Server  *hub.Service
	Devices *hub.DeviceManager
	Handler *hubnode.ReconnectHandler
}

// BuildNode builds the node of the config exactly once: the device manager
// with the handlers of the enabled managers, the RPC receiver and the
// connection to the master. The node is only usable when err is nil.
func BuildNode(cfg config.Service, opts ...Option) (*Node, error) {
	if cfg.MasterURL == "" {
		return nil, ErrNoMasterURL
	}
	o, err := resolve(cfg, opts)
	if err != nil {
		return nil, err
	}
	devices := hub.NewDeviceManager(o.logger, cfg)
	handler := &hubnode.ReconnectHandler{}
	rpcNode, err := hubnode.NewRPCNode(cfg, devices, handler)
	if err != nil {
		return nil, err
	}
	n := &Node{
		Devices: devices,
		Handler: handler,
		Server:  hub.NewService(hub.Deps{Logger: o.logger, Devices: devices, Config: cfg}),
	}
	addHandlers(devices, nodeHandlers(cfg, devicenode.NewMemoryDeviceStore(cfg.DeviceManager), o.logger))
	addHandlers(devices, o.handlers)
	nodeClient := endpointnode.New(endpointnode.Deps{
		Logger: o.logger, RPCNode: rpcNode, Devices: devices, Config: cfg, Managers: managerNames(cfg), Reconnect: handler,
	})
	handler.ConnectFunction = nodeClient.ConnectAndServe
	return n, addEndpoints(n.Server, manager.New(o.logger, devices, cfg), devices, nodeClient)
}

// Run connects to the master and runs the node until its server fails or ctx
// is done.
func (n *Node) Run(ctx context.Context) error {
	return errors.Join(n.Handler.HandleConnect(), n.Server.RunNode(ctx, n.Handler))
}
