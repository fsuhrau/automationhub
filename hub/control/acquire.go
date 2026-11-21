package control

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

func newToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Acquire locks the device for owner. Calling it again with the same owner
// renews and returns the existing lease.
func (s *Service) Acquire(dev device.Device, owner string) (Lease, error) {
	if owner == "" {
		return Lease{}, errors.New("lease owner is required")
	}
	lease, changed, err := s.acquire(dev, owner)
	if changed {
		// after the unlock of acquire
		s.notify(dev.DeviceID())
	}
	return lease, err
}

// acquire takes the lease under the lock; changed tells whether the device
// lost or got its controller.
func (s *Service) acquire(dev device.Device, owner string) (lease Lease, changed bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if l, ok := s.leases[dev.DeviceID()]; ok {
		if !now.After(l.ExpiresAt) {
			return s.renewOrRefuse(l, owner, now)
		}
		delete(s.leases, dev.DeviceID())
		unlockDevice(l.dev)
		changed = true
	}
	if err := dev.Lock(); err != nil {
		return Lease{}, changed, fmt.Errorf("%w: locked by a test run", ErrDeviceBusy)
	}
	return s.newLease(dev, owner, now), true, nil
}

// renewOrRefuse renews the active lease for its owner; for anyone else the
// device is busy.
func (s *Service) renewOrRefuse(l *leaseEntry, owner string, now time.Time) (Lease, bool, error) {
	if l.Owner != owner {
		return Lease{}, false, fmt.Errorf("%w: controlled by %s", ErrDeviceBusy, l.Owner)
	}
	s.touch(l, now)
	return l.Lease, false, nil
}

// newLease stores the lease of owner on the locked device.
func (s *Service) newLease(dev device.Device, owner string, now time.Time) Lease {
	l := &leaseEntry{
		Lease: Lease{DeviceID: dev.DeviceID(), Token: newToken(), Owner: owner, AcquiredAt: now},
		dev:   dev,
	}
	s.touch(l, now)
	s.leases[dev.DeviceID()] = l
	return l.Lease
}
