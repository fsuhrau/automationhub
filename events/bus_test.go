package events

import (
	"sync"
	"testing"
	"time"
)

type chanHandler chan int

func (c chanHandler) Handle(v int) { c <- v }

func recv(t *testing.T, c <-chan int) int {
	t.Helper()
	select {
	case v := <-c:
		return v
	case <-time.After(2 * time.Second):
		t.Fatal("handler not called")
		return 0
	}
}

func TestTriggerCallsEveryHandler(t *testing.T) {
	b := NewBus[int]()
	a, c := make(chanHandler, 1), make(chanHandler, 1)
	b.Register(a)
	b.Register(c)
	b.Trigger(7)
	if recv(t, a) != 7 || recv(t, c) != 7 {
		t.Fatal("wrong payload")
	}
}

func TestTriggerWithoutHandlersIsNoop(t *testing.T) {
	NewBus[int]().Trigger(1)
}

func TestBlockingHandlerDoesNotBlockOthers(t *testing.T) {
	b := NewBus[int]()
	release := make(chan struct{})
	b.Register(HandlerFunc[int](func(int) { <-release }))
	done := make(chan int, 1)
	b.Register(chanHandler(done))
	b.Trigger(3)
	if recv(t, done) != 3 {
		t.Fatal("wrong payload")
	}
	close(release)
}

func TestConcurrentRegisterAndTrigger(t *testing.T) {
	b := NewBus[int]()
	const n = 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			b.Register(HandlerFunc[int](func(int) {}))
		}()
		go func() {
			defer wg.Done()
			b.Trigger(1)
		}()
	}
	wg.Wait()
	if len(b.snapshot()) != n {
		t.Fatalf("handlers = %d, want %d", len(b.snapshot()), n)
	}
}
