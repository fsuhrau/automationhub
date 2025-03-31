package node

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/rpc"
)

type Service struct {
	logger         *logrus.Entry
	db             *gorm.DB
	devicesManager manager.Devices
	sseBroker      *sse.Broker
	cfg            config.Service
	reconnect      node.ConnectionHandler
	rpcNode        *node.RPCNode
	managers       []string
	c              *websocket.Conn
}

func New(logger *logrus.Logger, db *gorm.DB, rpcNode *node.RPCNode, dm manager.Devices, config config.Service, managers []string, reconnect node.ConnectionHandler) *Service {
	return &Service{
		logger:         logger.WithField("Service", "NodeAPI"),
		db:             db,
		devicesManager: dm,
		sseBroker:      sse.NewBroker(),
		cfg:            config,
		managers:       managers,
		reconnect:      reconnect,
		rpcNode:        rpcNode,
	}
}

func (s *Service) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	// register self as rpc server
	return rpc.Register(s.rpcNode)
}
