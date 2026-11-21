package control

import "fmt"

// Holder returns the active lease on the device, without its token.
func (s *Service) Holder(deviceID string) (Lease, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.leases[deviceID]
	if !ok || s.now().After(l.ExpiresAt) {
		return Lease{}, false
	}
	lease := l.Lease
	lease.Token = ""
	return lease, true
}

// Authorize checks the token against the device's lease and renews it.
func (s *Service) Authorize(deviceID, token string) error {
	if token == "" {
		return ErrLeaseRequired
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.leases[deviceID]
	now := s.now()
	if !ok || l.Token != token || now.After(l.ExpiresAt) {
		return ErrLeaseInvalid
	}
	s.touch(l, now)
	return nil
}

// TokenFor returns the lease token if owner holds the device. It lets callers
// that are identified by owner (e.g. an MCP session) use the token based API.
func (s *Service) TokenFor(deviceID, owner string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.leases[deviceID]
	if !ok || s.now().After(l.ExpiresAt) {
		return "", ErrLeaseRequired
	}
	if l.Owner != owner {
		return "", fmt.Errorf("%w: controlled by %s", ErrDeviceBusy, l.Owner)
	}
	return l.Token, nil
}
