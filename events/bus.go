package events

import "sync"

// Handler receives the payload of an event.
type Handler[T any] interface {
	Handle(T)
}

// HandlerFunc adapts a function to Handler.
type HandlerFunc[T any] func(T)

// Handle calls f(payload).
func (f HandlerFunc[T]) Handle(payload T) { f(payload) }

// Bus delivers payloads of type T to all registered handlers. It is safe for
// concurrent use.
type Bus[T any] struct {
	mu       sync.Mutex
	handlers []Handler[T]
}

// NewBus creates a bus without handlers.
func NewBus[T any]() *Bus[T] {
	return &Bus[T]{}
}

// Register adds a handler that receives every later trigger.
func (b *Bus[T]) Register(handler Handler[T]) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
}

// Trigger calls every registered handler in its own goroutine.
func (b *Bus[T]) Trigger(payload T) {
	for _, handler := range b.snapshot() {
		go handler.Handle(payload)
	}
}

func (b *Bus[T]) snapshot() []Handler[T] {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]Handler[T](nil), b.handlers...)
}
