package node

import (
	"fmt"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
)

func (h *Handler) Start() error {
	/* not needed */
	return nil
}

func (h *Handler) Stop() error {
	/* not needed */
	return nil
}

// GetDevice returns the device with the id and the node it lives on.
func (h *Handler) GetDevice(deviceID string) (manager.NodeIdentifier, device.Device, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for node, mng := range h.nodes {
		for _, devices := range mng {
			for _, dev := range devices {
				if dev.DeviceID() == deviceID {
					return node, dev, nil
				}
			}
		}
	}

	return "", nil, device.DeviceNotFoundError
}

func (h *Handler) StartDevice(deviceID string) error {
	node, _, err := h.GetDevice(deviceID)
	if err != nil {
		return err
	}

	return h.nodeManager.StartDevice(node, deviceID)
}

func (h *Handler) StopDevice(deviceID string) error {
	node, _, err := h.GetDevice(deviceID)
	if err != nil {
		return err
	}

	return h.nodeManager.StopDevice(node, deviceID)
}

func (h *Handler) GetDevices() ([]device.Device, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var devices []device.Device
	for _, mng := range h.nodes {
		for _, devs := range mng {
			devices = append(devices, devs...)
		}
	}
	return devices, nil
}

func (h *Handler) HasDevice(device2 device.Device) bool {
	_, _, err := h.GetDevice(device2.DeviceID())
	return err == nil
}

func (h *Handler) RegisterDevice(data device.RegisterData) (device.Device, error) {
	return nil, fmt.Errorf("register device not implemented")
}
