package master

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type nodeMaster struct {
	cfg config.Service
	dm  manager.Devices
	nm  manager.Nodes
	ch  node.ConnectionHandler

	wsUpgrader    websocket.Upgrader
	videoUpgrader websocket.Upgrader
}

//nolint:revive // argument-limit: the constructor signature is fixed by its callers in cmd/
func New(config config.Service, dm manager.Devices, nm manager.Nodes, ch node.ConnectionHandler) *nodeMaster {
	return &nodeMaster{
		cfg: config,
		dm:  dm,
		nm:  nm,
		ch:  ch,
		wsUpgrader: websocket.Upgrader{
			ReadBufferSize:  protocol.SocketFrameSize,
			WriteBufferSize: protocol.SocketFrameSize,
		},
		videoUpgrader: websocket.Upgrader{
			ReadBufferSize:  4 << 10,
			WriteBufferSize: 4 << 10,
		},
	}
}

func (s *nodeMaster) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	apiNode := auth.Group("/node")

	apiNode.POST("/log", s.Log)
	apiNode.POST("/action", s.Action)
	apiNode.POST("/devices", s.DevicesChanged)
	apiNode.GET("/connect", s.Connect)
	apiNode.GET("/video/:stream_id", s.Video)

	return nil
}
