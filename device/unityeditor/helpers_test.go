package unityeditor

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

const testVersion = "2022.3.10f1"

func epochTime() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }

// tickClock fires every After at once and moves the time by the wait, so
// polling loops run without sleeping. onTick sees the new time.
//
// The clock stops calling onTick when the test ends, so a background log
// follower cannot touch the test's files during cleanup.
type tickClock struct {
	*testkit.FakeClock
	onTick func(now time.Time)
	mu     sync.Mutex
	closed bool
}

func newTickClock(t *testing.T, onTick func(now time.Time)) *tickClock {
	c := &tickClock{FakeClock: testkit.NewFakeClock(epochTime()), onTick: onTick}
	t.Cleanup(func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.closed = true
	})
	return c
}

func (c *tickClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.Advance(d)
		if c.onTick != nil {
			c.onTick(c.Now())
		}
	}
	return c.FakeClock.After(0)
}

// fakeStorage is a storage.Device that records what the handler stores.
type fakeStorage struct {
	mu        sync.Mutex
	stored    models.Devices
	getErr    error
	newErr    error
	updateErr error
	created   []models.Device
	updated   []device.Device
}

func (s *fakeStorage) GetDevices(string) (models.Devices, error) { return s.stored, s.getErr }

func (s *fakeStorage) NewDevice(_ string, d models.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.created = append(s.created, d)
	return s.newErr
}

func (s *fakeStorage) GetDevice(string, string) (*models.Device, error) { return nil, nil }

func (s *fakeStorage) Update(_ string, d device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updated = append(s.updated, d)
	return s.updateErr
}

func (s *fakeStorage) updateCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.updated)
}

type testEnv struct {
	handler *Handler
	storage *fakeStorage
	runner  *testkit.FakeRunner
	clock   device.Clock
}

func newEnv(clock device.Clock) *testEnv {
	if clock == nil {
		clock = testkit.NewFakeClock(epochTime())
	}
	env := &testEnv{storage: &fakeStorage{}, runner: &testkit.FakeRunner{}, clock: clock}
	cfg := config.Manager{UnityPath: "/Unity", UnityBuildTarget: "iOS"}
	env.handler = NewHandlerWithDeps(cfg, env.storage, Deps{Runner: env.runner, Clock: clock, GOOS: "darwin"})
	return env
}

// project creates a Unity project dir with the editor version file.
func project(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	settings := filepath.Join(dir, "ProjectSettings")
	if err := os.MkdirAll(settings, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "m_EditorVersion: " + testVersion + "\n"
	if err := os.WriteFile(filepath.Join(settings, "ProjectVersion.txt"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// addEditor puts a device with the project dir into the handler.
func (e *testEnv) addEditor(id, projectDir string) *Device {
	dev := newDevice(infoOf(id), e.clock)
	dev.SetConfig(&models.Device{CustomParameter: []models.CustomParameter{{Key: "projectDir", Value: projectDir}}})
	e.handler.Add(dev)
	return dev
}

// launchLine is the command line StartDevice runs for the project.
func launchLine(dir string, at time.Time) string {
	return "/Unity/" + testVersion + "/Unity.app/Contents/MacOS/Unity" +
		" -buildTarget iOS -projectPath " + dir +
		" -overrideProfile automation -executeMethod AutomationLoader.LoadSceneAndConnect" +
		" -logFile " + filepath.Join(dir, "automation_hub_unity_"+itoa(at.UnixNano())+".log") +
		" -debugCodeOptimization --ump-channel-service-on-startup"
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

// fakeConn is a managerConn. Reads come from in, writes go to out.
type fakeConn struct {
	in       chan []byte
	out      chan []byte
	writeErr error
}

func newFakeConn() *fakeConn {
	return &fakeConn{in: make(chan []byte), out: make(chan []byte, sendQueueSize)}
}

func (c *fakeConn) ReadMessage() (int, []byte, error) {
	msg, ok := <-c.in
	if !ok {
		return 0, nil, io.EOF
	}
	return 1, msg, nil
}

func (c *fakeConn) WriteMessage(_ int, data []byte) error {
	c.out <- data
	return c.writeErr
}

var errBoom = errors.New("boom")

// failingProcess is a device.Process whose Kill fails.
type failingProcess struct{}

func (failingProcess) Wait() error { return nil }
func (failingProcess) Kill() error { return errBoom }

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

func infoOf(id string) generic.Info {
	return generic.Info{ID: id, Name: "editor " + id}
}
