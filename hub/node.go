package hub

import (
	"github.com/fsuhrau/automationhub/hub/manager"
)

type NodeStatus int

const (
	NodeStatusDisconnected NodeStatus = iota
	NodeStatusConnected
)

type Node struct {
	Identifier      string
	Hostname        string
	Address         string
	Port            int32
	OperationSystem string
	Path            string
	Managers        []string
	Status          NodeStatus
	Handler         manager.RPCClient
}
