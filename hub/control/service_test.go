package control

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
)

// fakeDevice implements the parts of device.Device the control service uses,
// plus device.InputCapable.
type fakeDevice struct {
	device.Device
	id string

	mu       sync.Mutex
	locked   bool
	unlocked int // Unlock calls (cancel the device context)
	released int // Release calls (keep the app connected)
	handlers []action.ActionHandler
	taps     [][2]int
	texts    []string
}

func newFakeDevice(id string) *fakeDevice { return &fakeDevice{id: id} }

func (d *fakeDevice) DeviceID() string                     { return d.id }
func (d *fakeDevice) IsAppConnected() bool                 { return true }
func (d *fakeDevice) Error(string, string, ...interface{}) {}
func (d *fakeDevice) Lock() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.locked {
		return errors.New("locked")
	}
	d.locked = true
	return nil
}
func (d *fakeDevice) Unlock() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.locked = false
	d.unlocked++
	return nil
}
func (d *fakeDevice) Release() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.locked = false
	d.released++
	return nil
}
func (d *fakeDevice) IsLocked() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.locked
}
func (d *fakeDevice) Tap(x, y int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.taps = append(d.taps, [2]int{x, y})
	return nil
}
func (d *fakeDevice) Swipe(int, int, int, int, time.Duration) error { return nil }
func (d *fakeDevice) InputText(text string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.texts = append(d.texts, text)
	return nil
}
func (d *fakeDevice) PressKey(string) error { return nil }
func (d *fakeDevice) AddActionHandler(h action.ActionHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = append(d.handlers, h)
}
func (d *fakeDevice) RemoveActionHandler(h action.ActionHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.handlers {
		if d.handlers[i] == h {
			d.handlers = append(d.handlers[:i:i], d.handlers[i+1:]...)
			return
		}
	}
}
func (d *fakeDevice) respond(resp *action.Response) {
	d.mu.Lock()
	hs := append([]action.ActionHandler(nil), d.handlers...)
	d.mu.Unlock()
	for _, h := range hs {
		h.OnActionResponse(d, resp)
	}
}

type devicesManager = manager.Devices

// fakeApp plays the Unity plugin: it answers scene graph and touch requests.
type fakeApp struct {
	devicesManager
	sceneXML string

	mu      sync.Mutex
	touched []string
}

func (m *fakeApp) SendAction(dev device.Device, a action.Interface) error {
	raw, err := a.Serialize()
	if err != nil {
		return err
	}
	var req action.Request
	_ = json.Unmarshal(raw, &req)
	fd := dev.(*fakeDevice)
	resp := &action.Response{ActionType: a.GetActionType(), Success: true}
	switch a.GetActionType() {
	case action.ActionType_GetSceneGraph:
		resp.Payload.Screenshot = &action.Screenshot{ContentType: action.ContentType_Xml, Sceengraph: []byte(m.sceneXML)}
	case action.ActionType_ElementTouch:
		m.mu.Lock()
		m.touched = append(m.touched, *req.Payload.Id)
		m.mu.Unlock()
	}
	go fd.respond(resp)
	return nil
}

func TestLeaseLifecycle(t *testing.T) {
	s := New(nil)
	dev := newFakeDevice("d1")

	lease, err := s.Acquire(dev, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if lease.Token == "" || !dev.IsLocked() {
		t.Fatal("expected token and locked device")
	}

	// same owner renews and gets the same lease
	again, err := s.Acquire(dev, "alice")
	if err != nil || again.Token != lease.Token {
		t.Fatalf("expected renewal of the same lease, got %v %v", again.Token, err)
	}

	// another owner is rejected
	if _, err := s.Acquire(dev, "bob"); !errors.Is(err, ErrDeviceBusy) {
		t.Fatalf("expected ErrDeviceBusy, got %v", err)
	}

	if err := s.Tap(dev, "", 1, 2); !errors.Is(err, ErrLeaseRequired) {
		t.Fatalf("expected ErrLeaseRequired, got %v", err)
	}
	if err := s.Tap(dev, "wrong", 1, 2); !errors.Is(err, ErrLeaseInvalid) {
		t.Fatalf("expected ErrLeaseInvalid, got %v", err)
	}
	if err := s.Tap(dev, lease.Token, 10, 20); err != nil {
		t.Fatal(err)
	}
	if len(dev.taps) != 1 || dev.taps[0] != [2]int{10, 20} {
		t.Fatalf("tap not forwarded: %v", dev.taps)
	}

	holder, ok := s.Holder("d1")
	if !ok || holder.Owner != "alice" || holder.Token != "" {
		t.Fatalf("unexpected holder %+v", holder)
	}

	if err := s.Release("d1", "wrong"); !errors.Is(err, ErrLeaseInvalid) {
		t.Fatalf("expected ErrLeaseInvalid, got %v", err)
	}
	if err := s.Release("d1", lease.Token); err != nil {
		t.Fatal(err)
	}
	if dev.IsLocked() || dev.released != 1 || dev.unlocked != 0 {
		t.Fatalf("expected release without cancelling the device context (released=%d unlocked=%d)", dev.released, dev.unlocked)
	}
}

func TestAcquireFailsWhileTestRunHoldsDevice(t *testing.T) {
	s := New(nil)
	dev := newFakeDevice("d1")
	_ = dev.Lock() // a test run locked it
	if _, err := s.Acquire(dev, "alice"); !errors.Is(err, ErrDeviceBusy) {
		t.Fatalf("expected ErrDeviceBusy, got %v", err)
	}
}

func TestIdleLeaseExpires(t *testing.T) {
	s := New(nil)
	now := time.Now()
	s.now = func() time.Time { return now }
	s.SetIdleTimeout(time.Minute)
	dev := newFakeDevice("d1")

	lease, err := s.Acquire(dev, "alice")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Minute)
	if err := s.Tap(dev, lease.Token, 0, 0); !errors.Is(err, ErrLeaseInvalid) {
		t.Fatalf("expected expired lease, got %v", err)
	}
	if n := s.ReapExpired(); n != 1 {
		t.Fatalf("expected one reaped lease, got %d", n)
	}
	if dev.IsLocked() {
		t.Fatal("expired lease must unlock the device")
	}
	// someone else can take it now
	if _, err := s.Acquire(dev, "bob"); err != nil {
		t.Fatal(err)
	}
}

func TestReleaseOwner(t *testing.T) {
	s := New(nil)
	d1, d2, d3 := newFakeDevice("d1"), newFakeDevice("d2"), newFakeDevice("d3")
	_, _ = s.Acquire(d1, "mcp:1")
	_, _ = s.Acquire(d2, "mcp:1")
	_, _ = s.Acquire(d3, "alice")
	if n := s.ReleaseOwner("mcp:1"); n != 2 {
		t.Fatalf("expected 2 released, got %d", n)
	}
	if d1.IsLocked() || d2.IsLocked() || !d3.IsLocked() {
		t.Fatal("wrong devices released")
	}
}

func TestTouchElementByXPath(t *testing.T) {
	app := &fakeApp{sceneXML: `<Root ID="1" Name="root" isVisible="1">` +
		`<Button ID="7" Name="Hidden" isVisible="0"/>` +
		`<Button ID="8" Name="Play" isVisible="1"/>` +
		`</Root>`}
	s := New(app)
	dev := newFakeDevice("d1")
	lease, err := s.Acquire(dev, "alice")
	if err != nil {
		t.Fatal(err)
	}

	if err := s.TouchElement(context.Background(), dev, lease.Token, ElementRef{XPath: "//Button[@Name='Play']"}); err != nil {
		t.Fatal(err)
	}
	if err := s.TouchElement(context.Background(), dev, lease.Token, ElementRef{XPath: "//Button[@Name='Nope']"}); err == nil {
		t.Fatal("expected error for unknown element")
	}
	if len(app.touched) != 1 || app.touched[0] != "8" {
		t.Fatalf("expected element 8 touched, got %v", app.touched)
	}

	xml, err := s.Hierarchy(context.Background(), dev, false)
	if err != nil || len(xml) == 0 {
		t.Fatalf("hierarchy failed: %v", err)
	}
}

func TestOnChangeNotifiesControllerChanges(t *testing.T) {
	s := New(nil)
	now := time.Now()
	s.now = func() time.Time { return now }
	s.SetIdleTimeout(time.Minute)
	var changes []string
	s.OnChange(func(id string) { changes = append(changes, id) })
	d1, d2 := newFakeDevice("d1"), newFakeDevice("d2")

	lease, _ := s.Acquire(d1, "alice")
	_, _ = s.Acquire(d1, "alice") // renewal: same controller, no change
	_, _ = s.Acquire(d1, "bob")   // rejected: no change
	if err := s.Release("d1", lease.Token); err != nil {
		t.Fatal(err)
	}
	_, _ = s.Acquire(d2, "mcp:1")
	s.ForceRelease("d2")
	_ = d2.Unlock()      // the unlock route unlocks the device itself
	s.ForceRelease("d2") // nothing to release
	_, _ = s.Acquire(d1, "mcp:1")
	s.ReleaseOwner("mcp:1")
	_, _ = s.Acquire(d2, "alice")
	now = now.Add(2 * time.Minute)
	s.ReapExpired()

	want := []string{"d1", "d1", "d2", "d2", "d1", "d1", "d2", "d2"}
	if len(changes) != len(want) {
		t.Fatalf("expected changes %v, got %v", want, changes)
	}
	for i := range want {
		if changes[i] != want[i] {
			t.Fatalf("expected changes %v, got %v", want, changes)
		}
	}
}

func TestOnChangeRunsOutsideTheLock(t *testing.T) {
	s := New(nil)
	dev := newFakeDevice("d1")
	var holder string
	s.OnChange(func(id string) {
		// would deadlock if called while the service lock is held
		if l, ok := s.Holder(id); ok {
			holder = l.Owner
		}
	})
	if _, err := s.Acquire(dev, "alice"); err != nil {
		t.Fatal(err)
	}
	if holder != "alice" {
		t.Fatalf("expected holder alice in callback, got %q", holder)
	}
}
