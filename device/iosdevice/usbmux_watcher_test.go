package iosdevice

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeListener replays events, then fails like a usbmuxd restart.
type fakeListener struct {
	events chan bool
	closed chan struct{}
	once   sync.Once
}

func newFakeListener(events ...bool) *fakeListener {
	l := &fakeListener{events: make(chan bool, len(events)), closed: make(chan struct{})}
	for _, e := range events {
		l.events <- e
	}
	close(l.events)
	return l
}

func (l *fakeListener) Next() (bool, string, error) {
	select {
	case e, ok := <-l.events:
		if !ok {
			return false, "", errors.New("usbmuxd went away")
		}
		return e, "serial", nil
	case <-l.closed:
		return false, "", errors.New("closed")
	}
}

func (l *fakeListener) Close() { l.once.Do(func() { close(l.closed) }) }

func TestWatchUsbmuxRefreshesOnEventsAndReconnects(t *testing.T) {
	var opens atomic.Int32
	open := func() (muxListener, error) {
		switch opens.Add(1) {
		case 1:
			return newFakeListener(true, false), nil // attach, detach, then it breaks
		case 2:
			return nil, errors.New("usbmuxd not running")
		default:
			// stays open until the watcher is stopped
			l := &fakeListener{events: make(chan bool), closed: make(chan struct{})}
			return l, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	var changes atomic.Int32
	var states []bool
	var mu sync.Mutex
	done := make(chan struct{})
	go func() {
		watchUsbmux(ctx, open, func() { changes.Add(1) }, func(up bool) {
			mu.Lock()
			states = append(states, up)
			mu.Unlock()
		})
		close(done)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for opens.Load() < 3 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if opens.Load() < 3 {
		t.Fatalf("watcher did not reconnect (opens=%d)", opens.Load())
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop")
	}

	// first connect + attach + detach, then one more on reconnect
	if got := changes.Load(); got != 4 {
		t.Fatalf("expected 4 refreshes, got %d", got)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(states) != 4 || !states[0] || states[1] || !states[2] || states[3] {
		t.Fatalf("expected up, down, up, down; got %v", states)
	}
}
