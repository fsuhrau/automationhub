package api

import (
	"fmt"

	"github.com/fsuhrau/automationhub/hub/sse"
)

// bridge publishes every payload of an event bus on an SSE channel. The
// channel name and the content sent to clients are derived from the payload.
type bridge[T any] struct {
	publisher sse.Publisher
	channel   func(T) string
	content   func(T) any
}

// Handle implements events.Handler.
func (b bridge[T]) Handle(payload T) {
	b.publisher.PublishEvent(sse.Event{Channel: b.channel(payload), Content: b.content(payload)})
}

// newBridge creates the handler that publishes events on the SSE channel.
func newBridge[T any](publisher sse.Publisher, channel func(T) string, content func(T) any) bridge[T] {
	return bridge[T]{publisher: publisher, channel: channel, content: content}
}

// fixedChannel always returns the same channel name.
func fixedChannel[T any](name string) func(T) string {
	return func(T) string { return name }
}

// idChannel formats the id of the payload into a channel name.
func idChannel[T any](format string, id func(T) uint) func(T) string {
	return func(payload T) string { return fmt.Sprintf(format, id(payload)) }
}

// payloadContent sends the payload itself.
func payloadContent[T any](payload T) any {
	return payload
}
