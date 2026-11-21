package scrcpy

import (
	"bytes"
	"sync"
)

// outputTailSize is how much of the server output is kept for errors.
const outputTailSize = 2048

// tailWriter keeps the last bytes written to it: the end of the scrcpy server
// output says why the server died. Its Write is safe for concurrent use.
type tailWriter struct {
	mu  sync.Mutex
	buf []byte
}

func (t *tailWriter) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(t.buf, p...)
	if extra := len(t.buf) - outputTailSize; extra > 0 {
		t.buf = t.buf[extra:]
	}
	return len(p), nil
}

// String is the tail without surrounding whitespace; a nil tail is empty.
func (t *tailWriter) String() string {
	if t == nil {
		return ""
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(bytes.TrimSpace(t.buf))
}
