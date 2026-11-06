// Package videostreamtest has a fake device.VideoStream for tests.
package videostreamtest

import (
	"io"
	"sync"

	"github.com/fsuhrau/automationhub/device"
)

// Stream returns the given packets, then blocks until it is closed. Injected
// input arrives on Inputs.
type Stream struct {
	meta    device.VideoMeta
	packets chan device.VideoPacket
	closed  chan struct{}
	once    sync.Once
	// Inputs receives the events of InjectInput.
	Inputs chan device.InputEvent
}

func New(meta device.VideoMeta, packets ...device.VideoPacket) *Stream {
	s := &Stream{meta: meta, packets: make(chan device.VideoPacket, len(packets)), closed: make(chan struct{}), Inputs: make(chan device.InputEvent, 16)}
	for _, p := range packets {
		s.packets <- p
	}
	return s
}

func (s *Stream) Meta() device.VideoMeta { return s.meta }

func (s *Stream) ReadPacket() (device.VideoPacket, error) {
	select {
	case p := <-s.packets:
		return p, nil
	case <-s.closed:
		return device.VideoPacket{}, io.EOF
	}
}

func (s *Stream) Close() error {
	s.once.Do(func() { close(s.closed) })
	return nil
}

// Closed is closed once the stream was closed.
func (s *Stream) Closed() <-chan struct{} { return s.closed }

// InjectInput implements device.VideoInput.
func (s *Stream) InjectInput(ev device.InputEvent) error {
	s.Inputs <- ev
	return nil
}
