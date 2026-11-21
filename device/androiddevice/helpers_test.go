package androiddevice

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

var errBoom = errors.New("boom")

// memoryStore is an in-memory storage.Device that records the updates.
type memoryStore struct {
	mu        sync.Mutex
	devices   models.Devices
	updates   []string
	created   []models.Device
	getErr    error
	createErr error
	updateErr error
}

func (s *memoryStore) GetDevices(string) (models.Devices, error) { return s.devices, s.getErr }

func (s *memoryStore) NewDevice(_ string, d models.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.created = append(s.created, d)
	s.devices = append(s.devices, &d)
	return s.createErr
}

func (s *memoryStore) GetDevice(_ string, id string) (*models.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, d := range s.devices {
		if d.DeviceIdentifier == id {
			return d, nil
		}
	}
	return nil, device.DeviceNotFoundError
}

func (s *memoryStore) Update(_ string, d device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updates = append(s.updates, fmt.Sprintf("%s=%s", d.DeviceID(), device.StateToString(d.DeviceState())))
	return s.updateErr
}

// testDeps are Deps with a scripted runner and a clock that never blocks.
func testDeps(runner *testkit.FakeRunner) Deps {
	return Deps{
		Runner:    runner,
		Clock:     &autoClock{},
		NewID:     func() string { return "tmp1" },
		AAPT:      func() (string, error) { return "aapt", nil },
		SessionID: func() int32 { return 0x1234 },
	}
}

// newTestDevice is a booted device S1 whose adb calls go to the runner.
func newTestDevice(runner *testkit.FakeRunner) *Device {
	d := newDevice(testDeps(runner), nil, "S1")
	d.State = device.StateBooted
	return d
}

// scriptAll scripts every line with the output "" and no error.
func scriptAll(runner *testkit.FakeRunner, lines ...string) {
	for _, l := range lines {
		runner.Script(l, nil, nil)
	}
}

func assertCommands(t *testing.T, runner *testkit.FakeRunner, want ...string) {
	t.Helper()
	if got := runner.Commands(); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("commands\n got %q\nwant %q", got, want)
	}
}

// pngBytes is a w x h PNG image.
func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// inTempDir runs the test in an empty working directory.
func inTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	return dir
}

func writeFile(t *testing.T, name string, data []byte) {
	t.Helper()
	if err := os.WriteFile(name, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
