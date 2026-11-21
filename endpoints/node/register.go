package node

import (
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type nodeInfo struct {
	identifier string
	hostname   string
	os         string
	path       string
	port       int32
	managers   []string
}

func (s *Service) registerNode(c *websocket.Conn, info nodeInfo) error {
	request := node.RegisterNodeRequest{
		Identifier:           info.identifier,
		Hostname:             info.hostname,
		OperationSystem:      info.os,
		EnvironmentVariables: info.path,
		Managers:             info.managers,
		Port:                 info.port,
	}
	data, err := proto.Marshal(&request)
	if err != nil {
		return err
	}

	return c.WriteMessage(websocket.BinaryMessage, data)
}
