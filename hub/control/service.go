// Package control provides interactive (live) device control on top of the
// device layer: native input, screenshots, logs, app lifecycle and app-level
// actions through the Unity plugin. It is the shared base for the REST API and
// the MCP server.
//
// Every state-changing call needs a lease. A lease locks the device (the same
// lock test runs use), so manual or agent control never collides with a
// running test. Leases expire after an idle timeout.
package control

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
)

const (
	DefaultIdleTimeout   = 10 * time.Minute
	DefaultActionTimeout = 20 * time.Second
	reapInterval         = 30 * time.Second
)

var (
	// ErrLeaseRequired is returned when a control call has no lease token.
	ErrLeaseRequired = errors.New("lock the device for control first")
	// ErrLeaseInvalid is returned when the token doesn't match the active lease.
	ErrLeaseInvalid = errors.New("device lease is invalid or expired")
	// ErrDeviceBusy is returned when the device is locked by a test run or by
	// another controller.
	ErrDeviceBusy = errors.New("device is in use")
	// ErrAppNotConnected is returned by app-level calls when no app is connected.
	ErrAppNotConnected = errors.New("no app connected to the device")
)

// Lease is an exclusive control session on one device.
type Lease struct {
	DeviceID   string    `json:"deviceId"`
	Token      string    `json:"token,omitempty"`
	Owner      string    `json:"owner"`
	AcquiredAt time.Time `json:"acquiredAt"`
	LastUsed   time.Time `json:"lastUsed"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type leaseEntry struct {
	Lease
	dev device.Device
}

// releaser is implemented by generic.Device: unlock without cancelling the
// device context, so a connected app stays connected.
type releaser interface {
	Release() error
}

type Service struct {
	dm            manager.Devices
	idleTimeout   time.Duration
	actionTimeout time.Duration
	now           func() time.Time

	mu     sync.Mutex
	leases map[string]*leaseEntry // by device id
}

func New(dm manager.Devices) *Service {
	return &Service{
		dm:            dm,
		idleTimeout:   DefaultIdleTimeout,
		actionTimeout: DefaultActionTimeout,
		now:           time.Now,
		leases:        make(map[string]*leaseEntry),
	}
}

// SetIdleTimeout changes how long an unused lease stays valid.
func (s *Service) SetIdleTimeout(d time.Duration) {
	s.mu.Lock()
	s.idleTimeout = d
	s.mu.Unlock()
}

// Run releases idle leases until ctx is done.
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(reapInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.ReapExpired()
		}
	}
}

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
	}
	return len(expired)
}

func newToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func unlockDevice(dev device.Device) {
	if r, ok := dev.(releaser); ok {
		_ = r.Release()
		return
	}
	_ = dev.Unlock()
}

// Acquire locks the device for owner. Calling it again with the same owner
// renews and returns the existing lease.
func (s *Service) Acquire(dev device.Device, owner string) (Lease, error) {
	if owner == "" {
		return Lease{}, errors.New("lease owner is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if l, ok := s.leases[dev.DeviceID()]; ok {
		if now.After(l.ExpiresAt) {
			delete(s.leases, dev.DeviceID())
			unlockDevice(l.dev)
		} else if l.Owner == owner {
			s.touch(l, now)
			return l.Lease, nil
		} else {
			return Lease{}, fmt.Errorf("%w: controlled by %s", ErrDeviceBusy, l.Owner)
		}
	}

	if err := dev.Lock(); err != nil {
		return Lease{}, fmt.Errorf("%w: locked by a test run", ErrDeviceBusy)
	}
	l := &leaseEntry{
		Lease: Lease{
			DeviceID:   dev.DeviceID(),
			Token:      newToken(),
			Owner:      owner,
			AcquiredAt: now,
		},
		dev: dev,
	}
	s.touch(l, now)
	s.leases[dev.DeviceID()] = l
	return l.Lease, nil
}

func (s *Service) touch(l *leaseEntry, now time.Time) {
	l.LastUsed = now
	l.ExpiresAt = now.Add(s.idleTimeout)
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
	return nil
}

// ForceRelease drops any lease on the device without unlocking it; used when
// the device is unlocked through the regular unlock route.
func (s *Service) ForceRelease(deviceID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.leases[deviceID]
	delete(s.leases, deviceID)
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
	}
	return len(released)
}

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
