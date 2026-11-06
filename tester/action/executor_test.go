package action

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
)

// fakeDevice implements only the methods the executor uses.
type fakeDevice struct {
	device.Device
	mu       sync.Mutex
	handlers []action.ActionHandler
}

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

func (d *fakeDevice) Error(source, format string, params ...interface{}) {}

func (d *fakeDevice) deliver(resp *action.Response) {
	d.mu.Lock()
	hs := append([]action.ActionHandler(nil), d.handlers...)
	d.mu.Unlock()
	for _, h := range hs {
		h.OnActionResponse(d, resp)
	}
}

func (d *fakeDevice) handlerCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.handlers)
}

// alias so the embedded field is not named Devices (clashes with the method)
type devicesManager = manager.Devices

// fakeManager calls onSend for every SendAction.
type fakeManager struct {
	devicesManager
	onSend func(dev device.Device, a action.Interface)
}

func (m *fakeManager) SendAction(dev device.Device, a action.Interface) error {
	if m.onSend != nil {
		go m.onSend(dev, a)
	}
	return nil
}

type fakeAction struct {
	processed atomic.Int32
}

func (a *fakeAction) GetActionType() action.ActionType { return action.ActionType_GetSceneGraph }
func (a *fakeAction) Serialize() ([]byte, error)       { return nil, nil }
func (a *fakeAction) ProcessResponse(*action.Response) error {
	a.processed.Add(1)
	return nil
}

func TestExecuteCompletesOnResponse(t *testing.T) {
	dev := &fakeDevice{}
	mgr := &fakeManager{onSend: func(d device.Device, a action.Interface) {
		dev.deliver(&action.Response{ActionType: a.GetActionType(), Success: true})
		// duplicate responses must neither block nor be processed twice
		dev.deliver(&action.Response{ActionType: a.GetActionType(), Success: true})
	}}
	a := &fakeAction{}
	if err := NewExecutor(mgr).Execute(context.Background(), dev, a, time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.processed.Load() != 1 {
		t.Fatalf("expected one processed response, got %d", a.processed.Load())
	}
	if dev.handlerCount() != 0 {
		t.Fatal("handler not removed")
	}
}

func TestExecuteTimeoutAndLateResponseDoesNotBlock(t *testing.T) {
	dev := &fakeDevice{}
	mgr := &fakeManager{onSend: func(d device.Device, a action.Interface) {}}
	e := NewExecutor(mgr)

	start := time.Now()
	err := e.Execute(context.Background(), dev, &fakeAction{}, 50*time.Millisecond)
	if !errors.Is(err, TimeoutError) {
		t.Fatalf("expected timeout, got %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("timeout took too long")
	}

	// a late response delivered directly to the executor must return immediately
	done := make(chan struct{})
	go func() {
		e.OnActionResponse(dev, &action.Response{ActionType: action.ActionType_GetSceneGraph})
		e.OnActionResponse(dev, nil)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("late response blocked")
	}
}

func TestExecuteContextCancel(t *testing.T) {
	dev := &fakeDevice{}
	mgr := &fakeManager{}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	err := NewExecutor(mgr).Execute(ctx, dev, &fakeAction{}, 5*time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestExecuteDisconnectReturnsError(t *testing.T) {
	dev := &fakeDevice{}
	mgr := &fakeManager{onSend: func(d device.Device, a action.Interface) {
		dev.deliver(nil)
	}}
	err := NewExecutor(mgr).Execute(context.Background(), dev, &fakeAction{}, time.Second)
	if !errors.Is(err, ErrDeviceDisconnected) {
		t.Fatalf("expected ErrDeviceDisconnected, got %v", err)
	}
}
