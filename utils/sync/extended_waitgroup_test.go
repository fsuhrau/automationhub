package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitWithTimeoutLateDoneDoesNotPanic(t *testing.T) {
	wg := NewExtendedWaitGroup(context.Background())
	wg.Add(1)
	if err := wg.WaitWithTimeout(10 * time.Millisecond); !errors.Is(err, TimeoutError) {
		t.Fatalf("expected timeout, got %v", err)
	}
	if !wg.IsCanceled() {
		t.Fatal("expected canceled after timeout")
	}
	wg.Done() // must not panic
	time.Sleep(10 * time.Millisecond)
}

func TestWaitWithTimeoutContextCancelLateDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := NewExtendedWaitGroup(ctx)
	wg.Add(1)
	cancel()
	if err := wg.WaitWithTimeout(time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled, got %v", err)
	}
	wg.Done() // used to panic with send on closed channel
	time.Sleep(10 * time.Millisecond)
}

func TestWaitUntilHonoursDeadline(t *testing.T) {
	wg := NewExtendedWaitGroup(context.Background())
	wg.Add(1)
	start := time.Now()
	err := wg.WaitUntil(time.Now().Add(50 * time.Millisecond))
	if !errors.Is(err, TimeoutError) {
		t.Fatalf("expected timeout, got %v", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("deadline not honoured, took %v", time.Since(start))
	}
}

func TestWaitUntilExtendedByUpdateUntil(t *testing.T) {
	wg := NewExtendedWaitGroup(context.Background())
	wg.Add(1)
	go func() {
		time.Sleep(20 * time.Millisecond)
		wg.UpdateUntil(time.Now().Add(5 * time.Second))
		time.Sleep(100 * time.Millisecond)
		wg.Done()
	}()
	if err := wg.WaitUntil(time.Now().Add(60 * time.Millisecond)); err != nil {
		t.Fatalf("expected completion after extension, got %v", err)
	}
}

func TestWaitCompletes(t *testing.T) {
	wg := NewExtendedWaitGroup(context.Background())
	wg.Add(2)
	go wg.Done()
	go wg.Done()
	if err := wg.Wait(); err != nil {
		t.Fatal(err)
	}
}
