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
	"errors"
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
	// after waits between two reads of a followed log
	after func(time.Duration) <-chan time.Time
	// reapEvery is how often Run releases the idle leases
	reapEvery time.Duration

	mu       sync.Mutex
	leases   map[string]*leaseEntry // by device id
	onChange func(deviceID string)
}

func New(dm manager.Devices) *Service {
	return &Service{
		dm:            dm,
		idleTimeout:   DefaultIdleTimeout,
		actionTimeout: DefaultActionTimeout,
		now:           time.Now,
		after:         time.After,
		reapEvery:     reapInterval,
		leases:        make(map[string]*leaseEntry),
	}
}

// SetIdleTimeout changes how long an unused lease stays valid.
func (s *Service) SetIdleTimeout(d time.Duration) {
	s.mu.Lock()
	s.idleTimeout = d
	s.mu.Unlock()
}

// OnChange registers a callback that runs when a device gets a new controller
// or loses its controller, e.g. to update the device pool. It is called
// outside the service lock.
func (s *Service) OnChange(fn func(deviceID string)) {
	s.mu.Lock()
	s.onChange = fn
	s.mu.Unlock()
}

func (s *Service) notify(deviceIDs ...string) {
	s.mu.Lock()
	fn := s.onChange
	s.mu.Unlock()
	if fn == nil {
		return
	}
	for _, id := range deviceIDs {
		fn(id)
	}
}

// Run releases idle leases until ctx is done.
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(s.reapEvery)
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

func (s *Service) touch(l *leaseEntry, now time.Time) {
	l.LastUsed = now
	l.ExpiresAt = now.Add(s.idleTimeout)
}
