package sse

import (
	"bytes"
	"context"
	"errors"
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

func TestServeHTTPSendsFirstByteAndKeepAlive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	b := NewBroker()
	b.keepAlive = 20 * time.Millisecond
	r := gin.New()
	r.GET("/sse", HeadersMiddleware(), b.ServeHTTP())
	srv := httptest.NewServer(r)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/sse", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("unexpected content type %q", resp.Header.Get("Content-Type"))
	}

	buf := make([]byte, 0, 256)
	chunk := make([]byte, 64)
	for !bytes.Contains(buf, []byte(": keep-alive")) {
		n, err := resp.Body.Read(chunk)
		if err != nil {
			t.Fatalf("read: %v (got %q)", err, buf)
		}
		buf = append(buf, chunk[:n]...)
	}
	if !bytes.HasPrefix(buf, []byte(": connected\n\n")) {
		t.Fatalf("stream must start with a comment, got %q", buf)
	}
}

// brokenWriter is a response writer that fails every write.
type brokenWriter struct{ *httptest.ResponseRecorder }

func (brokenWriter) Write([]byte) (int, error)       { return 0, errors.New("client gone") }
func (brokenWriter) WriteString(string) (int, error) { return 0, errors.New("client gone") }

func brokenContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(brokenWriter{httptest.NewRecorder()})
	c.Request = httptest.NewRequest(http.MethodGet, "/sse", nil)
	return c
}

func TestWriteCommentReportsAClientThatIsGone(t *testing.T) {
	if writeComment(brokenContext(), "connected") {
		t.Fatal("a write to a broken connection succeeded")
	}
}

func TestStreamEndsWhenAKeepAliveCannotBeWritten(t *testing.T) {
	b := NewBroker()
	b.keepAlive = time.Millisecond
	client := b.Subscribe()
	defer b.Unsubscribe(client)
	done := make(chan struct{})

	go func() { b.stream(brokenContext(), client); close(done) }()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("the stream of a client that is gone did not end")
	}
}

func TestSendEventDropsAnEventThatCannotBeEncoded(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	sendEvent(c, Event{Channel: "x", Content: make(chan int)})

	if rec.Body.Len() != 0 {
		t.Fatalf("wrote %q for an event that can't be encoded", rec.Body)
	}
}
