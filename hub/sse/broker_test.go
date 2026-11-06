package sse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestPublishDeliversToAllClients(t *testing.T) {
	b := NewBrokerWithBuffer(4)
	c1, c2 := b.Subscribe(), b.Subscribe()
	b.Publish(Event{Channel: "x", Content: 1})
	for _, c := range []*Client{c1, c2} {
		select {
		case ev := <-c.Events():
			if ev.Channel != "x" {
				t.Fatalf("unexpected channel %q", ev.Channel)
			}
		case <-time.After(time.Second):
			t.Fatal("event not delivered")
		}
	}
}

func TestSlowClientIsEvictedAndPublishNeverBlocks(t *testing.T) {
	b := NewBrokerWithBuffer(2)
	slow := b.Subscribe()
	fast := b.Subscribe()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			b.Publish(Event{Channel: "x", Content: i})
			// drain the fast client so it stays registered
			<-fast.Events()
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked on a slow client")
	}

	if b.ClientCount() != 1 {
		t.Fatalf("expected slow client to be evicted, got %d clients", b.ClientCount())
	}
	// slow channel must be closed after its buffered events
	n := 0
	for range slow.Events() {
		n++
	}
	if n != 2 {
		t.Fatalf("expected 2 buffered events for slow client, got %d", n)
	}
}

func TestUnsubscribeIsIdempotentAndConcurrentSafe(t *testing.T) {
	b := NewBrokerWithBuffer(1)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		c := b.Subscribe()
		wg.Add(3)
		go func() { defer wg.Done(); b.Unsubscribe(c) }()
		go func() { defer wg.Done(); b.Unsubscribe(c) }()
		go func() { defer wg.Done(); b.Publish(Event{Channel: "y"}) }()
	}
	wg.Wait()
	if b.ClientCount() != 0 {
		t.Fatalf("expected no clients, got %d", b.ClientCount())
	}
}

func TestServeHTTPStopsOnClientDisconnect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	b := NewBroker()
	r := gin.New()
	r.GET("/sse", HeadersMiddleware(), b.ServeHTTP())

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/sse", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	finished := make(chan struct{})
	go func() {
		r.ServeHTTP(w, req)
		close(finished)
	}()

	waitFor(t, func() bool { return b.ClientCount() == 1 })
	b.Publish(Event{Channel: "hello", Content: "world"})
	cancel()

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after client disconnect")
	}
	if b.ClientCount() != 0 {
		t.Fatalf("client not removed after disconnect")
	}
	// publishing after disconnect must not panic or block
	b.Publish(Event{Channel: "after"})
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met")
}
