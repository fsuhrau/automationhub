package testkit

import (
	"context"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
)

var _ manager.Devices = (*FakeDeviceManager)(nil)

// FakeDeviceManager is a manager.Devices that serves the devices added with
// Add and records its calls.
type FakeDeviceManager struct {
	recorder

	// Found is returned by FindDevice, Registered by RegisterDevice.
	Found      device.Device
	Registered device.Device

	byManager map[string][]device.Device
}

// Add makes d known under the manager name.
func (m *FakeDeviceManager) Add(managerName string, d device.Device) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byManager == nil {
		m.byManager = map[string][]device.Device{}
	}
	m.byManager[managerName] = append(m.byManager[managerName], d)
}

func (m *FakeDeviceManager) Run(_ context.Context, runSocketListener bool) error {
	return m.record("Run", runSocketListener)
}

func (m *FakeDeviceManager) Devices() (map[string][]device.Device, error) {
	err := m.record("Devices")
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string][]device.Device, len(m.byManager))
	for name, devs := range m.byManager {
		out[name] = append([]device.Device(nil), devs...)
	}
	return out, err
}

func (m *FakeDeviceManager) SendAction(d device.Device, a action.Interface) error {
	return m.record("SendAction", d, a)
}

// GetDevice returns the device with the id and the name of its manager, or
// nil and "" when it is unknown.
func (m *FakeDeviceManager) GetDevice(id string) (device.Device, string) {
	_ = m.record("GetDevice", id)
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, devs := range m.byManager {
		for _, d := range devs {
			if d.DeviceID() == id {
				return d, name
			}
		}
	}
	return nil, ""
}

func (m *FakeDeviceManager) Start(d device.Device) error { return m.record("Start", d) }
func (m *FakeDeviceManager) Stop(d device.Device) error  { return m.record("Stop", d) }
func (m *FakeDeviceManager) AddHandler(h device.Handler) { _ = m.record("AddHandler", h) }
func (m *FakeDeviceManager) StopObserver()               { _ = m.record("StopObserver") }

func (m *FakeDeviceManager) FindDevice(p *app.Parameter) device.Device {
	_ = m.record("FindDevice", p)
	return m.Found
}

func (m *FakeDeviceManager) RegisterDevice(data device.RegisterData) (device.Device, error) {
	return m.Registered, m.record("RegisterDevice", data)
}

// Managers returns the manager names that have devices.
func (m *FakeDeviceManager) Managers() []string {
	_ = m.record("Managers")
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.byManager))
	for name := range m.byManager {
		names = append(names, name)
	}
	return names
}
