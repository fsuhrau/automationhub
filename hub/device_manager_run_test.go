package hub

import (
	"context"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestRunStartsHandlersAndNotifiesInitialList(t *testing.T) {
	n := &countingNotifier{}
	dm := newTestManager(t, WithNotifierFactory(factoryFor(n)))
	h := newFakeHandler("android_device", testkit.NewFakeDevice("phone"))
	h.initErr = context.Canceled // only logged
	dm.AddHandler(h)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := dm.Run(ctx, true); err != nil {
		t.Fatal(err)
	}
	if h.notifier == nil {
		t.Fatal("handler did not get the change notifier")
	}
	if n.notifications() != 1 {
		t.Fatalf("initial notifications = %d, want 1", n.notifications())
	}
	h.notifier()
	if n.notifications() != 2 {
		t.Fatal("the handler callback must reach the injected notifier")
	}
	snap := n.list()
	if len(snap) != 1 || snap[0].DeviceID != "phone" {
		t.Fatalf("snapshot = %+v", snap)
	}
}

func TestRefreshLoopRefreshesOnSignalAndTick(t *testing.T) {
	n := &countingNotifier{}
	dm := newTestManager(t, WithNotifierFactory(factoryFor(n)))
	dm.refreshInterval = time.Hour
	h := newFakeHandler("ios_sim")
	h.listErr = context.DeadlineExceeded // refresh errors are only logged
	dm.AddHandler(h)
	ctx, cancel := context.WithCancel(context.Background())
	if err := dm.Run(ctx, false); err != nil {
		t.Fatal(err)
	}

	dm.TriggerRefresh()
	dm.TriggerRefresh() // a full channel must not block
	waitFor(t, "signal refresh", func() bool { return h.refreshCount() >= 1 })
	cancel()
}

func TestRefreshLoopTicks(t *testing.T) {
	n := &countingNotifier{}
	dm := newTestManager(t, WithNotifierFactory(factoryFor(n)))
	dm.refreshInterval = 5 * time.Millisecond
	h := newFakeHandler("ios_sim")
	dm.AddHandler(h)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := dm.Run(ctx, false); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "ticker refresh", func() bool { return h.refreshCount() >= 2 })
}

func TestDefaultNotifierFactoryUsesMasterConfig(t *testing.T) {
	dm := newTestManager(t)
	dm.masterUrl, dm.nodeIdentifier = "", ""
	dm.newNotifier(func() []*node.DeviceResponse { return nil }).Notify() // no master: no-op
	noopNotifier{}.Notify()
	dm.masterUrl, dm.nodeIdentifier = "master:1", "node-1"
	if dm.masterNotifierFactory(func() []*node.DeviceResponse { return nil }) == nil {
		t.Fatal("expected a notifier")
	}
}

func TestStopObserverMarksStop(t *testing.T) {
	dm := newTestManager(t)
	dm.StopObserver()
	if !dm.stop {
		t.Fatal("stop flag not set")
	}
}

func TestRunWithoutMasterIsNoopNotifier(t *testing.T) {
	dm := newTestManager(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := dm.Run(ctx, false); err != nil {
		t.Fatal(err)
	}
}

func TestNewDeviceManagerTakesTheNodeAuthToken(t *testing.T) {
	cfg := config.Service{MasterURL: "m:1", Identifier: "n1", Auth: config.Auth{Token: &config.Token{AuthToken: "secret"}}}
	dm := NewDeviceManager(quietLogger(), cfg)
	if dm.authToken == nil || *dm.authToken != "secret" || dm.masterUrl != "m:1" || dm.nodeIdentifier != "n1" {
		t.Fatalf("manager = %+v", dm)
	}
}
