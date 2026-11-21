package node

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	hubnode "github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func booted(id string) *hubnode.DeviceResponse {
	return &hubnode.DeviceResponse{Manager: "android_device", DeviceID: id, State: hubnode.DeviceState_Booted}
}

func connectedNodes() *testkit.FakeNodes {
	return &testkit.FakeNodes{Status: 1, Known: []manager.NodeIdentifier{"node-1"}}
}

func TestInit_RestoresStoredDevicesAndRefreshes(t *testing.T) {
	nodes := connectedNodes()
	nodes.Managed = pushed(booted("phone"))
	h, store := newHandlerFor(t, nodes)
	var first models.Node
	h.db.First(&first, "identifier = ?", "node-1")
	store.stored = models.Devices{
		{NodeID: first.ID, Manager: Manager, DeviceIdentifier: "phone", Name: "Pixel", OS: "Android", OSVersion: "14", TargetVersion: "34", PlatformType: models.PlatformTypeAndroid},
		{NodeID: first.ID, Manager: Manager, DeviceIdentifier: "tablet"},
	}

	if err := h.Init("master", "node", nil); err != nil {
		t.Fatal(err)
	}
	if h.init {
		t.Error("init flag must be reset")
	}
	devs, _ := h.GetDevices()
	if len(devs) != 2 {
		t.Fatalf("devices = %d", len(devs))
	}
	nodeID, phone, err := h.GetDevice("phone")
	if err != nil || nodeID != "node-1" {
		t.Fatalf("GetDevice = %v, %v", err, nodeID)
	}
	if phone.DeviceState() != device.StateBooted {
		t.Errorf("phone reported by the node must be booted, got %v", phone.DeviceState())
	}
	if _, tablet, _ := h.GetDevice("tablet"); tablet.DeviceState() != device.StateRemoteDisconnected {
		t.Errorf("tablet not reported must be remote disconnected, got %v", tablet.DeviceState())
	}
}

func TestRestoreDevice_FillsTheDeviceFromTheStoredModel(t *testing.T) {
	h, _ := newHandlerFor(t, connectedNodes())
	var n models.Node
	h.db.First(&n, "identifier = ?", "node-1")
	model := &models.Device{NodeID: n.ID, Manager: "android_device", DeviceIdentifier: "phone", Name: "Pixel", OS: "Android", OSVersion: "14", TargetVersion: "34", PlatformType: models.PlatformTypeAndroid}
	if err := h.restoreDevice(model); err != nil {
		t.Fatal(err)
	}
	nodeID, dev, _ := h.GetDevice("phone")
	nd := dev.(*NodeDevice)
	if nodeID != "node-1" || nd.DeviceName() != "Pixel" || nd.DeviceOSName() != "Android" || nd.DeviceOSVersion() != "14" ||
		nd.TargetVersion() != "34" || nd.PlatformType() != int(models.PlatformTypeAndroid) || nd.DeviceState() != device.StateUnknown || nd.GetConfig() != model {
		t.Fatalf("device = %+v", nd)
	}
	if got := h.nodes["node-1"]["android_device"]; len(got) != 1 {
		t.Errorf("stored under the model's manager, got %v", h.nodes)
	}
}

func TestInit_Errors(t *testing.T) {
	t.Run("storage fails", func(t *testing.T) {
		h, store := newHandlerFor(t, connectedNodes())
		store.getErr = errors.New("db down")
		if err := h.Init("", "", nil); err == nil || err.Error() != "db down" {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("node of a stored device is unknown", func(t *testing.T) {
		h, store := newHandlerFor(t, connectedNodes())
		store.stored = models.Devices{{NodeID: 999, DeviceIdentifier: "x"}}
		if err := h.Init("", "", nil); err == nil {
			t.Fatal("expected an error")
		}
		if h.init {
			t.Error("init flag must be reset on error")
		}
	})
	t.Run("refresh fails", func(t *testing.T) {
		h, _ := newHandlerFor(t, connectedNodes())
		h.nodes["ghost"] = map[string][]device.Device{} // no database row
		if err := h.Init("", "", nil); err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestAddDevice_KeepsDevicesPerNodeAndManager(t *testing.T) {
	h, _ := newHandlerFor(t, connectedNodes())
	a, b, c := testkit.NewFakeDevice("a"), testkit.NewFakeDevice("b"), testkit.NewFakeDevice("c")
	h.addDevice("node-1", "m1", a)
	h.addDevice("node-1", "m1", b)
	h.addDevice("node-1", "m2", c)
	if len(h.nodes["node-1"]["m1"]) != 2 || len(h.nodes["node-1"]["m2"]) != 1 {
		t.Fatalf("nodes = %v", h.nodes)
	}
}

func TestRefreshDevices_IsThrottledUnlessForced(t *testing.T) {
	nodes := connectedNodes()
	h, _ := newHandlerFor(t, nodes)
	if err := h.RefreshDevices(false); err != nil {
		t.Fatal(err)
	}
	calls := len(nodes.CallsTo("GetManagers"))
	if calls != 1 {
		t.Fatalf("first refresh asks the node, calls = %d", calls)
	}
	_ = h.RefreshDevices(false)
	if len(nodes.CallsTo("GetManagers")) != calls {
		t.Error("a refresh within UpdateInterval must not ask the nodes")
	}
	_ = h.RefreshDevices(true)
	if len(nodes.CallsTo("GetManagers")) != calls+1 {
		t.Error("a forced refresh must ask the nodes")
	}
}

func TestRefreshNode_AppliesTheNodesAnswer(t *testing.T) {
	nodes := connectedNodes()
	nodes.Managed = pushed(booted("phone"))
	h, _ := newHandlerFor(t, nodes)
	if err := h.refreshNode("node-1"); err != nil {
		t.Fatal(err)
	}
	if _, d, _ := h.GetDevice("phone"); d == nil {
		t.Fatal("device not applied")
	}
	if err := h.refreshNode("unknown"); err == nil {
		t.Error("an unknown node has no database row")
	}
}

func TestFetchNodeDevices(t *testing.T) {
	nodes := connectedNodes()
	h, _ := newHandlerFor(t, nodes)
	if got := h.fetchNodeDevices("node-1"); got == nil || len(got) != 0 {
		t.Errorf("a node without devices answers an empty map, got %#v", got)
	}
	nodes.FailWith("GetManagers", errors.New("timeout"))
	if got := h.fetchNodeDevices("node-1"); got != nil {
		t.Errorf("a node that fails to answer is nil, got %#v", got)
	}
	nodes.Status = 0
	if got := h.fetchNodeDevices("node-1"); got != nil {
		t.Errorf("a disconnected node is nil, got %#v", got)
	}
}

func TestUpdateNodeDevices_NilListMeansNoDevices(t *testing.T) {
	h, _ := newHandlerFor(t, connectedNodes())
	if err := h.UpdateNodeDevices("node-1", nil, 0); err != nil {
		t.Fatal(err)
	}
	if err := h.UpdateNodeDevices("node-1", nil, 0); err != nil {
		t.Fatal("pushes without sentAt are never dropped")
	}
}

func TestSyncNode_UnknownNodeFails(t *testing.T) {
	h, _ := newHandlerFor(t, connectedNodes())
	if err := h.syncNode("ghost", nil, time.Now().UTC()); err == nil {
		t.Fatal("expected an error")
	}
}

func TestSyncNode_KeepsStoredConfigAndLoadsMissingOne(t *testing.T) {
	nodes := connectedNodes()
	h, store := newHandlerFor(t, nodes)
	stored := &models.Device{DeviceIdentifier: "phone", Name: "stored"}
	store.devices["phone"] = stored
	h.addDevice("node-1", Manager, &NodeDevice{})
	if err := h.UpdateNodeDevices("node-1", pushed(booted("phone")), 1); err != nil {
		t.Fatal(err)
	}
	// the device was unknown to the node manager and is new: its config is the model created from the report
	_, dev, _ := h.GetDevice("phone")
	if dev.(*NodeDevice).GetConfig() == nil {
		t.Fatal("new device has no config")
	}

	// restored devices have no config yet: it is loaded from the storage
	restored := h.newNodeDevice("node-1")
	restored.ID = "tablet"
	store.devices["tablet"] = &models.Device{DeviceIdentifier: "tablet"}
	h.addDevice("node-1", Manager, restored)
	_ = h.UpdateNodeDevices("node-1", pushed(booted("phone"), booted("tablet")), 2)
	if restored.GetConfig() != store.devices["tablet"] {
		t.Error("config not loaded from the storage")
	}
}

func TestListeners_ApplyNodeEvents(t *testing.T) {
	nodes := connectedNodes()
	nodes.Managed = pushed(booted("fetched"))
	h, store := newHandlerFor(t, nodes)

	devicesChangedListener{h}.Handle(events.NodeDevicesChangedPayload{NodeIdentifier: "node-1", Devices: pushed(booted("pushed")), SentAt: 1})
	if _, d, _ := h.GetDevice("pushed"); d == nil {
		t.Error("pushed list not applied")
	}
	devicesChangedListener{h}.Handle(events.NodeDevicesChangedPayload{NodeIdentifier: "node-1"})
	if _, d, _ := h.GetDevice("fetched"); d == nil {
		t.Error("announced change not fetched over RPC")
	}
	devicesChangedListener{h}.Handle(events.NodeDevicesChangedPayload{NodeIdentifier: "unknown"}) // logs the failure

	h.lastPush["node-1"] = 5
	h.Handle(events.NodeConnectedPayload{NodeIdentifier: "node-1"})
	if _, ok := h.lastPush["node-1"]; ok {
		t.Error("a connecting node starts a new push order")
	}
	h.lastPush["node-1"] = 5
	nodes.Status = 0
	nodeDisconnectedListener{h}.Handle(events.NodeDisconnectedPayload{NodeIdentifier: "node-1"})
	if _, ok := h.lastPush["node-1"]; ok {
		t.Error("a lost node starts a new push order")
	}
	if s := store.state("pushed"); s != device.StateNodeDisconnected {
		t.Errorf("state = %v", s)
	}
}

func TestRegisterListeners_SubscribesToTheNodeEvents(t *testing.T) {
	h, _ := newHandlerFor(t, connectedNodes())
	h.registerListeners() // must not panic; the events registry is process wide
}

func TestRun_LoadsTheManagersOfConnectedNodes(t *testing.T) {
	nodes := connectedNodes()
	nodes.Managed = pushed(booted("phone"))
	h, _ := newHandlerFor(t, nodes)
	if err := h.Run(context.TODO(), false); err != nil {
		t.Fatal(err)
	}
	if len(h.nodes["node-1"]["android_device"]) != 1 {
		t.Fatalf("nodes = %v", h.nodes)
	}
	nodes.FailWith("GetManagers", errors.New("down"))
	delete(h.nodes, "node-1")
	_ = h.Run(context.TODO(), false)
	if _, ok := h.nodes["node-1"]; ok {
		t.Error("a node that fails to answer is skipped")
	}
}

func TestHandlerDevices(t *testing.T) {
	nodes := connectedNodes()
	h, _ := newHandlerFor(t, nodes)
	if h.Name() != Manager || h.Start() != nil || h.Stop() != nil {
		t.Error("name/start/stop")
	}
	if _, err := h.RegisterDevice(device.RegisterData{}); err == nil {
		t.Error("register device is not implemented")
	}
	if err := h.StartDevice("nope"); !errors.Is(err, device.DeviceNotFoundError) {
		t.Errorf("StartDevice(unknown) = %v", err)
	}
	if err := h.StopDevice("nope"); !errors.Is(err, device.DeviceNotFoundError) {
		t.Errorf("StopDevice(unknown) = %v", err)
	}
	if h.HasDevice(testkit.NewFakeDevice("phone")) {
		t.Error("unknown device")
	}

	h.addDevice("node-1", Manager, testkit.NewFakeDevice("phone"))
	if !h.HasDevice(testkit.NewFakeDevice("phone")) {
		t.Error("known device")
	}
	if err := h.StartDevice("phone"); err != nil {
		t.Error(err)
	}
	if err := h.StopDevice("phone"); err != nil {
		t.Error(err)
	}
	if got := nodes.CallsTo("StartDevice"); len(got) != 1 || got[0].Args[0] != manager.NodeIdentifier("node-1") || got[0].Args[1] != "phone" {
		t.Errorf("StartDevice calls = %v", got)
	}
	if len(nodes.CallsTo("StopDevice")) != 1 {
		t.Error("StopDevice not forwarded")
	}
}

func TestSync_StorageFailuresAreLoggedNotFatal(t *testing.T) {
	h, store := newHandlerFor(t, connectedNodes())
	store.writeErr = errors.New("disk full")
	if err := h.UpdateNodeDevices("node-1", pushed(booted("phone")), 1); err != nil {
		t.Fatalf("a storage failure must not fail the sync: %v", err)
	}
	if _, d, _ := h.GetDevice("phone"); d == nil {
		t.Fatal("device must be known although it could not be stored")
	}
}
