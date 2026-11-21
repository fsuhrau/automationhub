package node

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/rpc"
	"sync"
	"sync/atomic"
)

// Service is the node side of the connection to the master: it dials the
// master and serves the node's RPC methods over that connection.
type Service struct {
	logger         *logrus.Entry
	devicesManager manager.Devices
	cfg            config.Service
	reconnect      node.ConnectionHandler
	rpcNode        *node.RPCNode
	managers       []string
	connMu         sync.Mutex
	c              *websocket.Conn
	dialing        atomic.Bool
	rpcServer      *rpc.Server
}

// Deps are what the node service is built on.
type Deps struct {
	Logger    *logrus.Logger
	RPCNode   *node.RPCNode
	Devices   manager.Devices
	Config    config.Service
	Managers  []string
	Reconnect node.ConnectionHandler
}

// New builds the node service.
func New(deps Deps) *Service {
	return &Service{
		logger:         deps.Logger.WithField("Service", "NodeAPI"),
		devicesManager: deps.Devices,
		cfg:            deps.Config,
		managers:       deps.Managers,
		reconnect:      deps.Reconnect,
		rpcNode:        deps.RPCNode,
		rpcServer:      rpc.NewServer(),
	}
}

func (s *Service) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	// register self as rpc server
	return s.rpcServer.Register(s.rpcNode)
}
