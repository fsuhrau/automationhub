package generic

import (
	"sort"
	"sync"

	"github.com/fsuhrau/automationhub/device"
)

// Registry is the mutex-safe set of devices a handler manages, keyed by device
// id. The zero value is ready to use. Platform handlers embed it for
// GetDevices and HasDevice.
type Registry[D device.Device] struct {
	mu      sync.RWMutex
	devices map[string]D
}

// Add stores the device, replacing one with the same id.
func (r *Registry[D]) Add(d D) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.devices == nil {
		r.devices = map[string]D{}
	}
	r.devices[d.DeviceID()] = d
}

// Remove drops the device with the id and reports whether it existed.
func (r *Registry[D]) Remove(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.devices[id]
	delete(r.devices, id)
	return ok
}

// Get returns the device with the id.
func (r *Registry[D]) Get(id string) (D, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.devices[id]
	return d, ok
}

// Has reports whether a device with the id exists.
func (r *Registry[D]) Has(id string) bool {
	_, ok := r.Get(id)
	return ok
}

// HasDevice reports whether the registry holds the device, matched by id.
func (r *Registry[D]) HasDevice(d device.Device) bool {
	return r.Has(d.DeviceID())
}

// List returns the devices ordered by id.
func (r *Registry[D]) List() []D {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]D, 0, len(r.devices))
	for _, d := range r.devices {
		list = append(list, d)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].DeviceID() < list[j].DeviceID() })
	return list
}

// GetDevices returns the devices as device.Device, as device.Handler wants it.
func (r *Registry[D]) GetDevices() ([]device.Device, error) {
	list := r.List()
	devices := make([]device.Device, len(list))
	for i, d := range list {
		devices[i] = d
	}
	return devices, nil
}
