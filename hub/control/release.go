package control

import "github.com/fsuhrau/automationhub/device"

// ReapExpired releases all leases that were idle longer than the timeout.
func (s *Service) ReapExpired() int {
	now := s.now()
	var expired []*leaseEntry
	s.mu.Lock()
	for id, l := range s.leases {
		if now.After(l.ExpiresAt) {
			expired = append(expired, l)
			delete(s.leases, id)
		}
	}
	s.mu.Unlock()
	for _, l := range expired {
		unlockDevice(l.dev)
		s.notify(l.DeviceID)
	}
	return len(expired)
}

// Release ends the lease identified by token.
func (s *Service) Release(deviceID, token string) error {
	s.mu.Lock()
	l, ok := s.leases[deviceID]
	if !ok || l.Token != token || token == "" {
		s.mu.Unlock()
		return ErrLeaseInvalid
	}
	delete(s.leases, deviceID)
	s.mu.Unlock()
	unlockDevice(l.dev)
	s.notify(deviceID)
	return nil
}

// ForceRelease drops any lease on the device without unlocking it; used when
// the device is unlocked through the regular unlock route.
func (s *Service) ForceRelease(deviceID string) bool {
	s.mu.Lock()
	_, ok := s.leases[deviceID]
	delete(s.leases, deviceID)
	s.mu.Unlock()
	if ok {
		s.notify(deviceID)
	}
	return ok
}

// ReleaseOwner releases every lease held by owner, e.g. when an MCP session ends.
func (s *Service) ReleaseOwner(owner string) int {
	var released []*leaseEntry
	s.mu.Lock()
	for id, l := range s.leases {
		if l.Owner == owner {
			released = append(released, l)
			delete(s.leases, id)
		}
	}
	s.mu.Unlock()
	for _, l := range released {
		unlockDevice(l.dev)
		s.notify(l.DeviceID)
	}
	return len(released)
}

func unlockDevice(dev device.Device) {
	if r, ok := dev.(releaser); ok {
		_ = r.Release()
		return
	}
	_ = dev.Unlock()
}
