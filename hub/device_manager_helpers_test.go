package hub

import (
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/sirupsen/logrus"
)

func quietLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l
}

func newTestManager(t testing.TB, opts ...DeviceManagerOption) *DeviceManager {
	t.Helper()
	return NewDeviceManager(quietLogger(), config.Service{}, opts...)
}

// fakeHandler is a device.Handler over a fixed device list; it also takes a
// change notifier like the platform handlers do.
type fakeHandler struct {
	name     string
	devices  []device.Device
	registry device.RegisterData
	listErr  error
	initErr  error
	notifier func()

	mu        sync.Mutex
	refreshed int
	started   []string
	stopped   []string
}

func newFakeHandler(name string, devs ...device.Device) *fakeHandler {
	return &fakeHandler{name: name, devices: devs}
}

func (h *fakeHandler) Name() string                       { return h.name }
func (h *fakeHandler) Init(string, string, *string) error { return h.initErr }
func (h *fakeHandler) Start() error                       { return h.initErr }
func (h *fakeHandler) Stop() error                        { return nil }
func (h *fakeHandler) StartDevice(id string) error {
	h.started = append(h.started, id)
	return nil
}
func (h *fakeHandler) StopDevice(id string) error {
	h.stopped = append(h.stopped, id)
	return nil
}
func (h *fakeHandler) GetDevices() ([]device.Device, error) { return h.devices, h.listErr }
func (h *fakeHandler) RefreshDevices(bool) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.refreshed++
	return h.listErr
}
func (h *fakeHandler) refreshCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.refreshed
}
func (h *fakeHandler) HasDevice(d device.Device) bool {
	for _, own := range h.devices {
		if own.DeviceID() == d.DeviceID() {
			return true
		}
	}
	return false
}
func (h *fakeHandler) RegisterDevice(data device.RegisterData) (device.Device, error) {
	h.registry = data
	return testkit.NewFakeDevice(data.DeviceID), nil
}
func (h *fakeHandler) SetChangeNotifier(notify func()) { h.notifier = notify }

// countingNotifier counts Notify calls.
type countingNotifier struct {
	mu    sync.Mutex
	count int
	list  func() []*node.DeviceResponse
}

func (n *countingNotifier) Notify() {
	n.mu.Lock()
	n.count++
	n.mu.Unlock()
}

func (n *countingNotifier) notifications() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.count
}

func factoryFor(n *countingNotifier) NotifierFactory {
	return func(snapshot func() []*node.DeviceResponse) ChangeNotifier {
		n.list = snapshot
		return n
	}
}

// fakeAction is an action.Interface.
type fakeAction struct{ err error }

func (a *fakeAction) GetActionType() action.ActionType { return action.ActionType_Custom }
func (a *fakeAction) Serialize() ([]byte, error)       { return []byte("payload"), a.err }
func (a *fakeAction) ProcessResponse(*action.Response) error {
	return errors.New("unused")
}
