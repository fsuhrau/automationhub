package control

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestAcquireAfterAnExpiredLeaseReportsTheLostController(t *testing.T) {
	s := New(nil)
	now := time.Now()
	s.now = func() time.Time { return now }
	var changed []string
	s.OnChange(func(id string) { changed = append(changed, id) })
	dev := testkit.NewFakeDevice("d1")
	if _, err := s.Acquire(dev, "alice"); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * DefaultIdleTimeout)
	dev.FailWith("Lock", errors.New("a test run took it"))

	_, err := s.Acquire(dev, "bob")

	if !errors.Is(err, ErrDeviceBusy) {
		t.Fatalf("err %v", err)
	}
	if len(changed) != 2 {
		t.Fatalf("changes %v: alice got and lost the device", changed)
	}
}

func TestTokenForTheOwnerOfTheLease(t *testing.T) {
	s := New(nil)
	now := time.Now()
	s.now = func() time.Time { return now }
	dev := testkit.NewFakeDevice("d1")
	lease, _ := s.Acquire(dev, "alice")

	if got, err := s.TokenFor("d1", "alice"); err != nil || got != lease.Token {
		t.Fatalf("owner: %q, %v", got, err)
	}
	if _, err := s.TokenFor("d1", "bob"); !errors.Is(err, ErrDeviceBusy) {
		t.Fatalf("other owner: %v", err)
	}
	if _, err := s.TokenFor("d2", "alice"); !errors.Is(err, ErrLeaseRequired) {
		t.Fatalf("no lease: %v", err)
	}
	now = now.Add(2 * DefaultIdleTimeout)
	if _, err := s.TokenFor("d1", "alice"); !errors.Is(err, ErrLeaseRequired) {
		t.Fatalf("expired lease: %v", err)
	}
	if _, ok := s.Holder("d1"); ok {
		t.Fatal("an expired lease has no holder")
	}
}

func TestReleaseUnlocksADeviceWithoutARelease(t *testing.T) {
	s := New(nil)
	dev := testkit.NewFakeDevice("d1")
	lease, _ := s.Acquire(dev, "alice")

	if err := s.Release("d1", lease.Token); err != nil {
		t.Fatal(err)
	}

	if len(dev.CallsTo("Unlock")) != 1 || dev.IsLocked() {
		t.Fatalf("calls %v", dev.Calls())
	}
}

func TestRunReleasesIdleLeasesUntilTheContextEnds(t *testing.T) {
	s := New(nil)
	now := time.Now()
	s.now = func() time.Time { return now }
	s.reapEvery = time.Millisecond
	dev := testkit.NewFakeDevice("d1")
	_, _ = s.Acquire(dev, "alice")
	now = now.Add(2 * DefaultIdleTimeout)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() { s.Run(ctx); close(done) }()

	deadline := time.After(5 * time.Second)
	for dev.IsLocked() {
		select {
		case <-deadline:
			t.Fatal("the idle lease was not reaped")
		case <-time.After(time.Millisecond):
		}
	}
	cancel()
	<-done
}

// uploader is a node device: the bundle goes to the node first.
type uploader struct {
	*testkit.FakeDevice
	nodes *testkit.FakeNodes
}

func (u uploader) GetNodeID() manager.NodeIdentifier { return "node-1" }
func (u uploader) NodeManager() manager.Nodes        { return u.nodes }

func TestInstallAppUploadsTheBundleToTheNodeFirst(t *testing.T) {
	s := New(nil)
	node := uploader{FakeDevice: testkit.NewFakeDevice("d1"), nodes: &testkit.FakeNodes{}}
	lease, _ := s.Acquire(node, "alice")
	target := Target{Device: node, Token: lease.Token}

	if err := s.InstallApp(context.Background(), target, &app.Parameter{}); err != nil {
		t.Fatal(err)
	}
	if len(node.nodes.CallsTo("UploadApp")) != 1 || len(node.CallsTo("InstallApp")) != 1 {
		t.Fatalf("calls %v / %v", node.nodes.Calls(), node.Calls())
	}

	node.nodes.FailWith("UploadApp", errors.New("offline"))
	if err := s.InstallApp(context.Background(), target, &app.Parameter{}); err == nil {
		t.Fatal("an upload error must stop the install")
	}
}

func TestElementCallsOnAnAppThatFailsOrIsGone(t *testing.T) {
	app := &fakeApp{sceneXML: `<Root ID="1" Name="root" isVisible="1"><Button ID="8" Name="Play" isVisible="1"/></Root>`}
	s := New(app)
	dev := newFakeDevice("d1")
	lease, _ := s.Acquire(dev, "alice")
	target := Target{Device: dev, Token: lease.Token}
	byID := ElementRef{ID: "8"}

	if err := s.SetElementValue(context.Background(), target, ElementChange{Ref: byID, Attr: "text", Value: "hi"}); err != nil {
		t.Fatalf("set value: %v", err)
	}
	app.reject = true
	if err := s.SetElementValue(context.Background(), target, ElementChange{Ref: byID, Attr: "text"}); err == nil {
		t.Error("a rejected value must be an error")
	}
	if err := s.TouchElement(context.Background(), target, byID); err == nil {
		t.Error("a rejected touch must be an error")
	}
	if err := s.TouchElement(context.Background(), target, ElementRef{}); err == nil {
		t.Error("an element needs an id or an xpath")
	}
	wrong := Target{Device: dev, Token: "wrong"}
	if err := s.SetElementValue(context.Background(), wrong, ElementChange{Ref: byID}); !errors.Is(err, ErrLeaseInvalid) {
		t.Errorf("without a lease: %v", err)
	}
	if err := s.TouchElement(context.Background(), wrong, byID); !errors.Is(err, ErrLeaseInvalid) {
		t.Errorf("touch without a lease: %v", err)
	}
}

func TestAppCallsNeedAConnectedApp(t *testing.T) {
	s := New(&fakeApp{})
	dev := testkit.NewFakeDevice("d1") // no app connected

	_, _, _, err := s.AppScreenshot(context.Background(), dev)
	_, hierarchyErr := s.Hierarchy(context.Background(), dev, false)

	if !errors.Is(err, ErrAppNotConnected) || !errors.Is(hierarchyErr, ErrAppNotConnected) {
		t.Fatalf("errors %v / %v", err, hierarchyErr)
	}
	if _, _, _, err := s.Screenshot(&bareDevice{id: "p"}); !errors.Is(err, device.ErrUnsupported) {
		t.Fatalf("screenshot of a bare device: %v", err)
	}
}

func TestElementCallsReportAnUnknownElementAndAnAppThatIsGone(t *testing.T) {
	s := New(&fakeApp{sceneXML: `<Root ID="1" Name="root" isVisible="1"/>`})
	connected := newFakeDevice("d1")
	lease, _ := s.Acquire(connected, "alice")
	unknown := ElementRef{XPath: "//Button[@Name='Nope']"}
	gone := testkit.NewFakeDevice("d2") // no app connected
	goneLease, _ := s.Acquire(gone, "alice")

	if err := s.SetElementValue(context.Background(), Target{connected, lease.Token}, ElementChange{Ref: unknown}); err == nil {
		t.Error("an unknown element must be an error")
	}
	byID := ElementRef{ID: "8"}
	if err := s.SetElementValue(context.Background(), Target{gone, goneLease.Token}, ElementChange{Ref: byID}); !errors.Is(err, ErrAppNotConnected) {
		t.Errorf("set value without an app: %v", err)
	}
	if err := s.TouchElement(context.Background(), Target{gone, goneLease.Token}, byID); !errors.Is(err, ErrAppNotConnected) {
		t.Errorf("touch without an app: %v", err)
	}
}
