package node

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

// coreNodes is a node manager without the optional device roles.
type coreNodes struct{ manager.Nodes }

func TestNodeDevice_WithoutNodeRolesTheCallsAreUnsupported(t *testing.T) {
	fake := &testkit.FakeNodes{}
	d := newForwardingDevice(t, coreNodes{fake})

	if err := d.StartRecording("/x"); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("StartRecording = %v, want one wrapping device.ErrUnsupported", err)
	}
	if err := d.StopRecording(); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("StopRecording = %v, want one wrapping device.ErrUnsupported", err)
	}
	if _, _, _, err := d.GetScreenshot(); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("GetScreenshot = %v, want one wrapping device.ErrUnsupported", err)
	}
}

func TestNodeDevice_WithoutNodeRolesFeaturesAndScriptsAreSkipped(t *testing.T) {
	fake := &testkit.FakeNodes{Feature: true}
	d := newForwardingDevice(t, coreNodes{fake})

	if d.HasFeature("lua") {
		t.Error("HasFeature must be false without a features role")
	}
	d.Execute("cmd")
	d.RunNativeScript([]byte("s"))
	if calls := fake.Calls(); len(calls) != 0 {
		t.Errorf("nothing may reach the node manager, got %v", calls)
	}
}
