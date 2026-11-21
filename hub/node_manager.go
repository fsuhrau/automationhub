package hub

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var ErrNodeNotFound = fmt.Errorf("node not found")
var ErrNodeNotConnected = fmt.Errorf("node not connected")

const (
	NodeConnectionTimeout   = 5 * time.Minute
	NodeHealthCheckInterval = 5 * time.Second
)

type NodeManager struct {
	db         *gorm.DB
	mu         sync.RWMutex
	nodes      map[manager.NodeIdentifier]Node
	log        *logrus.Entry
	rendezvous *videostream.Rendezvous
}

func NewNodeManager(logger *logrus.Logger, db *gorm.DB) *NodeManager {
	return &NodeManager{log: logger.WithFields(logrus.Fields{
		"prefix": "dm",
	}),
		db:         db,
		nodes:      make(map[manager.NodeIdentifier]Node),
		rendezvous: videostream.NewRendezvous(),
	}
}

// Rendezvous pairs the video streams the nodes open with the requests that
// wait for them. Everything of the master process shares this one.
func (nm *NodeManager) Rendezvous() *videostream.Rendezvous {
	return nm.rendezvous
}

func (nm *NodeManager) Run(ctx context.Context) {
	nm.log.Debugf("Start NodeManager")
	go func() {
		ticker := time.NewTicker(NodeHealthCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				nm.log.Infof("Stop NodeManager")
				return
			case <-ticker.C:
				nm.healthCheck()
			}
		}
	}()
}

// healthCheck pings all connected nodes without holding the lock during the
// RPC call and marks unreachable nodes as disconnected.
func (nm *NodeManager) healthCheck() {
	nm.mu.RLock()
	handlers := make(map[manager.NodeIdentifier]manager.RPCClient, len(nm.nodes))
	for k, n := range nm.nodes {
		if n.Handler != nil {
			handlers[k] = n.Handler
		}
	}
	nm.mu.RUnlock()

	for k, h := range handlers {
		if err := h.Ping(); err != nil {
			nm.mu.Lock()
			// only mark disconnected if the node was not re-registered meanwhile
			disconnected := false
			if n, ok := nm.nodes[k]; ok && n.Handler == h {
				n.Status = NodeStatusDisconnected
				n.Handler = nil
				nm.nodes[k] = n
				disconnected = true
			}
			nm.mu.Unlock()
			if disconnected {
				nm.log.Infof("node %s disconnected: %v", k, err)
				events.NodeDisconnected.Trigger(events.NodeDisconnectedPayload{NodeIdentifier: k})
			}
		}
	}
}

func (nm *NodeManager) GetNodes() []manager.NodeIdentifier {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	var nodes []manager.NodeIdentifier
	for k := range nm.nodes {
		nodes = append(nodes, k)
	}
	return nodes
}

func (nm *NodeManager) GetManagers(nodeIdentifier manager.NodeIdentifier) (map[string][]device.Device, error) {
	handler, err := nm.getHandler(nodeIdentifier)
	if err == nil {
		managers, err := handler.GetDevices()
		if err != nil {
			nm.log.Errorf("fail to fetch managers from node '%s': %v", nodeIdentifier, err)
			return nil, err
		}
		return managers, nil
	}
	return nil, fmt.Errorf("node '%s' not found", nodeIdentifier)
}

func (nm *NodeManager) IsNodeKnown(nodeIdentifier manager.NodeIdentifier) bool {
	var dbNode models.Node
	if err := nm.db.First(&dbNode, "identifier = ?", nodeIdentifier).Error; err == gorm.ErrRecordNotFound {
		return false
	}
	return true
}

//nolint:revive // argument-limit: fixed by manager.NodeRegistry
func (nm *NodeManager) RegisterNode(nodeIdentifier manager.NodeIdentifier, client manager.RPCClient, hostname, address, operationSystem, environment string, port int32, managers []string) {
	/*
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			fmt.Println(err)
		}
	*/
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.nodes[nodeIdentifier] = Node{
		Identifier:      string(nodeIdentifier),
		Hostname:        hostname,
		Address:         address,
		Port:            port,
		OperationSystem: operationSystem,
		Path:            environment,
		Managers:        managers,
		Status:          NodeStatusConnected,
		Handler:         client,
	}
}

func (nm *NodeManager) getNode(nodeIdentifier manager.NodeIdentifier) (*Node, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	if v, ok := nm.nodes[nodeIdentifier]; ok {
		return &v, nil
	}

	return nil, ErrNodeNotFound
}

func (nm *NodeManager) getHandler(nodeIdentifier manager.NodeIdentifier) (manager.RPCClient, error) {
	n, err := nm.getNode(nodeIdentifier)
	if err != nil {
		return nil, err
	}

	if n.Handler == nil {
		return nil, ErrNodeNotConnected
	}
	return n.Handler, nil
}

func (nm *NodeManager) GetStatus(nodeIdentifier manager.NodeIdentifier) (int, error) {
	n, err := nm.getNode(nodeIdentifier)
	if err != nil {
		return 0, err
	}

	return int(n.Status), nil
}
