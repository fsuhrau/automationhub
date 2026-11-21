package unity

import (
	"context"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/tester/base"
)

// nodeHost is a device that lives on a node.
type nodeHost interface {
	GetNodeID() manager.NodeIdentifier
	NodeManager() manager.Nodes
}

// Prepare uploads the app once to every node that hosts one of the devices.
func (s *strategy) Prepare(ctx context.Context, params app.Parameter, devices []base.DeviceMap) {
	nodes := usedNodes(devices)
	if len(nodes) == 0 {
		return
	}
	s.runner.LogInfo("Upload new App to Nodes")
	for id, mng := range nodes {
		if err := mng.UploadApp(ctx, id, &params); err != nil {
			s.runner.LogError("Upload app to nodes failed: %v", err)
		}
	}
}

// usedNodes returns the unique nodes of the devices.
func usedNodes(devices []base.DeviceMap) map[manager.NodeIdentifier]manager.Nodes {
	nodes := make(map[manager.NodeIdentifier]manager.Nodes)
	for _, d := range devices {
		host, ok := d.Device.(nodeHost)
		if !ok {
			continue
		}
		if _, found := nodes[host.GetNodeID()]; !found {
			nodes[host.GetNodeID()] = host.NodeManager()
		}
	}
	return nodes
}
