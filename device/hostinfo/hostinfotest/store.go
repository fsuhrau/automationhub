// Package hostinfotest holds the test fakes shared by the macos and web
// platform tests.
package hostinfotest

import (
	"sync"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Store is a storage.Device that records the devices a handler reports.
type Store struct {
	mu      sync.Mutex
	err     error
	updated []device.Device
}

var _ storage.Device = (*Store)(nil)

// NewStore returns a Store whose Update fails with err (nil: succeeds).
func NewStore(err error) *Store { return &Store{err: err} }

// Updated returns the devices passed to Update in order.
func (s *Store) Updated() []device.Device {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]device.Device(nil), s.updated...)
}

func (s *Store) Update(_ string, d device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updated = append(s.updated, d)
	return s.err
}

func (s *Store) GetDevices(string) (models.Devices, error) { return nil, nil }

func (s *Store) NewDevice(string, models.Device) error { return nil }

func (s *Store) GetDevice(_, _ string) (*models.Device, error) {
	return nil, device.DeviceNotFoundError
}
