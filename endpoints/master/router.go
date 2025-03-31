package master

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/gin-gonic/gin"
)

type nodeMaster struct {
	cfg config.Service
	dm  manager.Devices
	nm  manager.Nodes
	ch  node.ConnectionHandler
}

func New(config config.Service, dm manager.Devices, nm manager.Nodes, ch node.ConnectionHandler) *nodeMaster {
	return &nodeMaster{
		cfg: config,
		dm:  dm,
		nm:  nm,
		ch:  ch,
	}
}

func (s *nodeMaster) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	apiNode := auth.Group("/node")

	apiNode.POST("/log", s.Log)
	apiNode.POST("/action", s.Action)
	apiNode.GET("/connect", s.Connect)

	return nil
}
