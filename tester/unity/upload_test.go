package unity

import (
	"context"
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/tester/base"
)

// uploads records the uploads of a node manager.
type uploads struct {
	manager.Nodes
	nodes []manager.NodeIdentifier
	err   error
}

func (u *uploads) UploadApp(_ context.Context, node manager.NodeIdentifier, _ *app.Parameter) error {
	u.nodes = append(u.nodes, node)
	return u.err
}

// nodeDevice is a fake device that lives on a node.
type nodeDevice struct {
	*testkit.FakeDevice
	node manager.NodeIdentifier
	mng  manager.Nodes
}

func (d *nodeDevice) GetNodeID() manager.NodeIdentifier { return d.node }
func (d *nodeDevice) NodeManager() manager.Nodes        { return d.mng }

func TestPrepareUploadsTheAppOncePerNode(t *testing.T) {
	f := newStrategyFixture(t)
	up := &uploads{}
	devices := []base.DeviceMap{
		{Device: &nodeDevice{FakeDevice: testkit.NewFakeDevice("1"), node: "n1", mng: up}},
		{Device: &nodeDevice{FakeDevice: testkit.NewFakeDevice("2"), node: "n1", mng: up}},
		{Device: &nodeDevice{FakeDevice: testkit.NewFakeDevice("3"), node: "n2", mng: up}},
		{Device: testkit.NewFakeDevice("local")},
	}
	f.strategy.Prepare(context.Background(), app.Parameter{}, devices)

	if len(up.nodes) != 2 || !f.logged("log", "Upload new App to Nodes") {
		t.Fatalf("uploads %v", up.nodes)
	}
}

func TestPrepareLogsAFailedUpload(t *testing.T) {
	f := newStrategyFixture(t)
	up := &uploads{err: errors.New("offline")}
	dev := &nodeDevice{FakeDevice: testkit.NewFakeDevice("1"), node: "n1", mng: up}
	f.strategy.Prepare(context.Background(), app.Parameter{}, []base.DeviceMap{{Device: dev}})

	if !f.logged("error", "Upload app to nodes failed: offline") {
		t.Fatal("failed upload not logged")
	}
}

func TestPrepareWithoutNodesDoesNothing(t *testing.T) {
	f := newStrategyFixture(t)
	f.strategy.Prepare(context.Background(), app.Parameter{}, []base.DeviceMap{{Device: testkit.NewFakeDevice("local")}})

	if f.logged("log", "Upload") {
		t.Fatal("uploaded without nodes")
	}
}
