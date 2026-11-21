package iossim

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

const (
	udid1 = "AAAA-1111"
	udid2 = "BBBB-2222"
)

// listJSON is `simctl list devices --json` with two iOS simulators and a tvOS one.
const listJSON = `{"devices":{
"com.apple.CoreSimulator.SimRuntime.iOS-17-0":[
 {"state":"Booted","isAvailable":true,"name":"iPhone 15","udid":"AAAA-1111"},
 {"state":"Shutdown","isAvailable":true,"name":"iPhone 15 Pro","udid":"BBBB-2222"}],
"com.apple.CoreSimulator.SimRuntime.tvOS-17-0":[
 {"state":"Shutdown","isAvailable":true,"name":"Apple TV","udid":"CCCC-3333"}],
"unknown.runtime":[{"state":"Booted","name":"Ghost","udid":"DDDD-4444"}]}}`

const listLine = "xcrun simctl list devices --json"

var errBoom = errors.New("boom")

// memoryStore is an in-memory storage.Device that records its calls.
type memoryStore struct {
	mu      sync.Mutex
	devices models.Devices
	updates []string
	getErr  error
	newErr  error
	updErr  error
}

func (s *memoryStore) GetDevices(string) (models.Devices, error) { return s.devices, s.getErr }

func (s *memoryStore) NewDevice(_ string, d models.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices = append(s.devices, &d)
	return s.newErr
}

func (s *memoryStore) GetDevice(string, string) (*models.Device, error) {
	return nil, device.DeviceNotFoundError
}

func (s *memoryStore) Update(_ string, d device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updates = append(s.updates, d.DeviceID()+"="+device.StateToString(d.DeviceState()))
	return s.updErr
}

func newTestDevice(runner *testkit.FakeRunner, keepRunning bool) *Device {
	return NewDevice(generic.Info{ID: udid1}, Deps{Runner: runner, KeepAppRunning: keepRunning})
}

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
