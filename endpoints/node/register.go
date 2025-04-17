package node

import (
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func (rpc *Service) RegisterNode(identifier, hostname, os, path string, port int32, managers []string) error {
	request := node.RegisterNodeRequest{
		Identifier:           identifier,
		Hostname:             hostname,
		OperationSystem:      os,
		EnvironmentVariables: path,
		Managers:             managers,
		Port:                 port,
	}
	data, err := proto.Marshal(&request)
	if err != nil {
		return err
	}

	return rpc.c.WriteMessage(websocket.BinaryMessage, data)
}
