package runs

import (
	"context"
	"testing"
	"time"
)

func TestTrackerLifecycle(t *testing.T) {
	finished := make(chan uint, 1)
	tr := NewTracker(func(id uint) { finished <- id })

	if state, _ := tr.Status(1); state != RunUnknown {
		t.Fatalf("expected unknown, got %s", state)
	}
	tr.Started(1)
	if state, _ := tr.Status(1); state != RunRunning {
		t.Fatalf("expected running, got %s", state)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, _, err := tr.Wait(ctx, 1); err == nil {
		t.Fatal("expected timeout while running")
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		tr.Finished(1, RunResult{Success: true, Succeeded: 3})
	}()
	state, result, err := tr.Wait(context.Background(), 1)
	if err != nil || state != RunDone || result == nil || result.Succeeded != 3 {
		t.Fatalf("unexpected wait result %s %+v %v", state, result, err)
	}
	select {
	case id := <-finished:
		if id != 1 {
			t.Fatalf("onFinish called with %d", id)
		}
	case <-time.After(time.Second):
		t.Fatal("onFinish not called")
	}
}
