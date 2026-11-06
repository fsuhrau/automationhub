package sse

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// DefaultClientBuffer is the number of events buffered per client before the
// client is considered too slow and gets evicted.
const DefaultClientBuffer = 256

// KeepAliveInterval is how often an idle stream gets a comment line, so
// proxies don't time out the connection.
var KeepAliveInterval = 25 * time.Second

// Client is a single SSE subscriber.
type Client struct {
	events chan Event
	once   sync.Once
}

// Events returns the channel the client receives events on. It is closed when
// the client is removed from the broker (unsubscribed or evicted).
func (c *Client) Events() <-chan Event {
	return c.events
}

// Broker fans out published events to all subscribed clients.
// Publishing never blocks: a client whose buffer is full is evicted, its
// channel is closed and the HTTP stream ends, so the browser reconnects.
type Broker struct {
	mu         sync.RWMutex
	clients    map[*Client]struct{}
	bufferSize int
}

func NewBroker() *Broker {
	return NewBrokerWithBuffer(DefaultClientBuffer)
}

func NewBrokerWithBuffer(bufferSize int) *Broker {
	if bufferSize < 1 {
		bufferSize = 1
	}
	return &Broker{
		clients:    make(map[*Client]struct{}),
		bufferSize: bufferSize,
	}
}

// Subscribe registers a new client.
func (b *Broker) Subscribe() *Client {
	c := &Client{events: make(chan Event, b.bufferSize)}
	b.mu.Lock()
	b.clients[c] = struct{}{}
	n := len(b.clients)
	b.mu.Unlock()
	log.Printf("SSE client added. %d registered clients", n)
	return c
}

// Unsubscribe removes the client. It is safe to call multiple times and
// concurrently with Publish.
func (b *Broker) Unsubscribe(c *Client) {
	c.once.Do(func() {
		b.mu.Lock()
		delete(b.clients, c)
		n := len(b.clients)
		// close under the write lock so no Publish can send on a closed channel
		close(c.events)
		b.mu.Unlock()
		log.Printf("SSE client removed. %d registered clients", n)
	})
}

// ClientCount returns the number of registered clients.
func (b *Broker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// Publish sends the event to all clients without blocking.
func (b *Broker) Publish(event Event) {
	var slow []*Client
	b.mu.RLock()
	for c := range b.clients {
		select {
		case c.events <- event:
		default:
			slow = append(slow, c)
		}
	}
	b.mu.RUnlock()

	for _, c := range slow {
		log.Printf("SSE client too slow, evicting")
		b.Unsubscribe(c)
	}
}

// PublishEvent implements Publisher.
func (b *Broker) PublishEvent(event Event) {
	b.Publish(event)
}

// ServeHTTP streams events to the requesting client until the request is
// cancelled or the client is evicted.
func (b *Broker) ServeHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := b.Subscribe()
		defer b.Unsubscribe(client)

		ctx := c.Request.Context()
		// send a first byte right away: some proxies (e.g. the Vite dev
		// server) hold back the headers until body data arrives, so the
		// browser's EventSource would never see the stream open
		if _, err := io.WriteString(c.Writer, ": connected\n\n"); err != nil {
			return
		}
		c.Writer.Flush()

		keepAlive := time.NewTicker(KeepAliveInterval)
		defer keepAlive.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-keepAlive.C:
				if _, err := io.WriteString(c.Writer, ": keep-alive\n\n"); err != nil {
					return
				}
				c.Writer.Flush()
			case msg, ok := <-client.Events():
				if !ok {
					return
				}
				c.SSEvent(msg.Channel, msg.Content)
				c.Writer.Flush()
			}
		}
	}
}
