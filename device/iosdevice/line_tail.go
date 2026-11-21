package iosdevice

import (
	"strings"
	"sync"
)

// lineTail keeps the last lines of the xcodebuild output for errors.
type lineTail struct {
	mu    sync.Mutex
	max   int
	lines []string
}

func newLineTail(n int) *lineTail { return &lineTail{max: n} }

func (t *lineTail) add(line string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lines = append(t.lines, line)
	if len(t.lines) > t.max {
		t.lines = t.lines[len(t.lines)-t.max:]
	}
}

func (t *lineTail) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return strings.Join(t.lines, "\n")
}
