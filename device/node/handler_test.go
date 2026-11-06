package node

import (
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	hubnode "github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// fakeNodes is a node manager whose nodes never answer RPC; the tests check
// that pushed device lists are applied without asking the node.
type fakeNodes struct {
	manager.Nodes
	mu        sync.Mutex
	connected map[manager.NodeIdentifier]bool
	rpcCalls  int
}

func (f *fakeNodes) GetStatus(n manager.NodeIdentifier) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.connected[n] {
		return 1, nil
	}
	return 0, nil
}

func (f *fakeNodes) GetNodes() []manager.NodeIdentifier {
	f.mu.Lock()
	defer f.mu.Unlock()
	var nodes []manager.NodeIdentifier
	for n := range f.connected {
		nodes = append(nodes, n)
	}
	return nodes
}

func (f *fakeNodes) GetManagers(manager.NodeIdentifier) (map[string][]device.Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rpcCalls++
	return map[string][]device.Device{}, nil
}

type fakeDeviceStore struct {
	mu      sync.Mutex
	devices map[string]*models.Device
	states  map[string]device.State
}

func (s *fakeDeviceStore) GetDevices(string) (models.Devices, error) { return nil, nil }

func (s *fakeDeviceStore) NewDevice(_ string, dev models.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[dev.DeviceIdentifier] = &dev
	return nil
}

func (s *fakeDeviceStore) GetDevice(_ string, id string) (*models.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.devices[id], nil
}

func (s *fakeDeviceStore) Update(_ string, dev device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[dev.DeviceID()] = dev.DeviceState()
	return nil
}

func (s *fakeDeviceStore) state(id string) device.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.states[id]
}

func pushed(devices ...*hubnode.DeviceResponse) map[string][]device.Device {
	return hubnode.FromDevicesResponse(devices)
}

func newTestHandler(t *testing.T) (*Handler, *fakeNodes, *fakeDeviceStore) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Node{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Create(&models.Node{Identifier: "node-1"}).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	nodes := &fakeNodes{connected: map[manager.NodeIdentifier]bool{"node-1": true}}
	store := &fakeDeviceStore{devices: map[string]*models.Device{}, states: map[string]device.State{}}
	return NewHandler(logrus.New(), store, nodes, db), nodes, store
}

func TestUpdateNodeDevicesAppliesPushWithoutRPC(t *testing.T) {
	h, nodes, store := newTestHandler(t)

	err := h.UpdateNodeDevices("node-1", pushed(
		&hubnode.DeviceResponse{Manager: "android_device", DeviceID: "phone", Name: "Pixel", State: hubnode.DeviceState_Unauthorized},
	), 1)
	if err != nil {
		t.Fatalf("apply push: %v", err)
	}
	if _, ok := store.devices["phone"]; !ok {
		t.Fatal("new device was not stored")
	}
	if s := store.state("phone"); s != device.StateUnauthorized {
		t.Fatalf("expected unauthorized, got %v", s)
	}

	// authorized on the phone
	_ = h.UpdateNodeDevices("node-1", pushed(
		&hubnode.DeviceResponse{Manager: "android_device", DeviceID: "phone", Name: "Pixel", State: hubnode.DeviceState_Booted},
	), 2)
	if s := store.state("phone"); s != device.StateBooted {
		t.Fatalf("expected booted, got %v", s)
	}

	// unplugged: the node no longer reports it
	_ = h.UpdateNodeDevices("node-1", pushed(), 3)
	if s := store.state("phone"); s != device.StateRemoteDisconnected {
		t.Fatalf("expected remote disconnected, got %v", s)
	}

	if nodes.rpcCalls != 0 {
		t.Fatalf("push must not ask the node over RPC (%d calls)", nodes.rpcCalls)
	}
	devs, _ := h.GetDevices()
	if len(devs) != 1 {
		t.Fatalf("expected one known device, got %d", len(devs))
	}
}

func TestUpdateNodeDevicesDropsStalePush(t *testing.T) {
	h, _, store := newTestHandler(t)

	_ = h.UpdateNodeDevices("node-1", pushed(
		&hubnode.DeviceResponse{Manager: "android_device", DeviceID: "phone", State: hubnode.DeviceState_Booted},
	), 20)
	// an older push that arrived late
	_ = h.UpdateNodeDevices("node-1", pushed(), 10)
	if s := store.state("phone"); s != device.StateBooted {
		t.Fatalf("stale push was applied, state %v", s)
	}

	// after a reconnect the order starts over (the node clock may have moved)
	h.forgetPushes("node-1")
	_ = h.UpdateNodeDevices("node-1", pushed(), 10)
	if s := store.state("phone"); s != device.StateRemoteDisconnected {
		t.Fatalf("push after reconnect was dropped, state %v", s)
	}
}

func TestUpdateNodeDevicesIgnoresDisconnectedNode(t *testing.T) {
	h, nodes, store := newTestHandler(t)
	nodes.connected["node-1"] = false

	_ = h.UpdateNodeDevices("node-1", pushed(
		&hubnode.DeviceResponse{Manager: "android_device", DeviceID: "phone", State: hubnode.DeviceState_Booted},
	), 1)
	if len(store.devices) != 0 {
		t.Fatal("push of a disconnected node was applied")
	}
}

func TestRefreshDevicesMarksOnlyOfflineNodeDevices(t *testing.T) {
	h, nodes, store := newTestHandler(t)
	if err := h.db.Create(&models.Node{Identifier: "node-2"}).Error; err != nil {
		t.Fatal(err)
	}
	nodes.connected["node-2"] = true

	_ = h.UpdateNodeDevices("node-1", pushed(&hubnode.DeviceResponse{Manager: "android_device", DeviceID: "a", State: hubnode.DeviceState_Booted}), 1)
	_ = h.UpdateNodeDevices("node-2", pushed(&hubnode.DeviceResponse{Manager: "android_device", DeviceID: "b", State: hubnode.DeviceState_Booted}), 1)

	// a push from node-2 doesn't touch the devices of node-1
	_ = h.UpdateNodeDevices("node-2", pushed(&hubnode.DeviceResponse{Manager: "android_device", DeviceID: "b", State: hubnode.DeviceState_Booted}), 2)
	if s := store.state("a"); s != device.StateBooted {
		t.Fatalf("device of another node changed to %v", s)
	}

	// node-1 drops: its devices are node disconnected after the refresh
	nodes.mu.Lock()
	nodes.connected["node-1"] = false
	nodes.mu.Unlock()
	if err := h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}
	if s := store.state("a"); s != device.StateNodeDisconnected {
		t.Fatalf("expected node disconnected, got %v", s)
	}
}
