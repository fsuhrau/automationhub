package hub

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/sirupsen/logrus"
)

// coreRPC is the connection of a node without the optional device roles.
type coreRPC struct{ manager.RPCClient }

func newCoreNodeManager() (*NodeManager, *fakeRPC) {
	nm := NewNodeManager(logrus.New(), nil)
	rpcClient := &fakeRPC{}
	nm.RegisterNode("n1", coreRPC{rpcClient}, "host", "1.2.3.4", "linux", "env", 9, nil)
	return nm, rpcClient
}

func TestRoleForwarding_UnsupportedByTheConnection(t *testing.T) {
	nm, rpcClient := newCoreNodeManager()

	if err := nm.StartRecording("n1", "d", "p"); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("StartRecording = %v, want one wrapping device.ErrUnsupported", err)
	}
	if err := nm.StopRecording("n1", "d"); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("StopRecording = %v, want one wrapping device.ErrUnsupported", err)
	}
	if _, _, _, err := nm.GetScreenshot("n1", "d"); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("GetScreenshot = %v, want one wrapping device.ErrUnsupported", err)
	}
	if nm.HasFeature("n1", "d", "f") {
		t.Error("HasFeature must be false without the role")
	}
	nm.Execute("n1", "d", "x")
	nm.RunNativeScript("n1", "d", []byte("s"))
	if len(rpcClient.calls) != 0 {
		t.Errorf("no call may reach the connection: %v", rpcClient.calls)
	}
}
