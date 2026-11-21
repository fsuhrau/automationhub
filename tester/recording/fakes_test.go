package recording

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

var errDisk = errors.New("disk full")

// fakeClock moves by step on every Now; After fires at once.
type fakeClock struct {
	mu      sync.Mutex
	now     time.Time
	step    time.Duration
	afters  int
	onAfter func(n int)
	// fireStop lets the stop timeout fire.
	fireStop bool
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), step: 10 * time.Millisecond}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := c.now
	c.now = c.now.Add(c.step)
	return t
}

// After fires at once, except for the stop timeout, which only fires when
// fireStop is set.
func (c *fakeClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	c.afters++
	n, hook := c.afters, c.onAfter
	c.mu.Unlock()
	if hook != nil {
		hook(n)
	}
	ch := make(chan time.Time, 1)
	if d != stopTimeout || c.fireStop {
		ch <- time.Time{}
	}
	return ch
}

// memFS keeps the files in memory and fails on demand.
type memFS struct {
	mu        sync.Mutex
	files     map[string]*memFile
	removed   []string
	createErr func(name string) error
	// failWriteAt makes the n-th Write (1-based, over all files) fail.
	failWriteAt int
	writes      int
	writeAtErr  error
	seekErr     error
	closeErr    error
}

func newMemFS() *memFS { return &memFS{files: map[string]*memFile{}} }

func (m *memFS) Create(name string) (File, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		if err := m.createErr(name); err != nil {
			return nil, err
		}
	}
	f := &memFile{fs: m, name: name}
	m.files[name] = f
	return f, nil
}

func (m *memFS) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removed = append(m.removed, name)
	delete(m.files, name)
	return nil
}

func (m *memFS) file(name string) *memFile {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.files[name]
}

// names of the files, in no order.
func (m *memFS) names() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []string
	for n := range m.files {
		out = append(out, n)
	}
	return out
}

type memFile struct {
	fs     *memFS
	name   string
	data   []byte
	pos    int64
	closed bool
}

func (f *memFile) Write(p []byte) (int, error) {
	f.fs.mu.Lock()
	f.fs.writes++
	fail := f.fs.failWriteAt > 0 && f.fs.writes >= f.fs.failWriteAt
	f.fs.mu.Unlock()
	if fail {
		return 0, errDisk
	}
	return f.writeAt(p, f.pos), nil
}

func (f *memFile) writeAt(p []byte, off int64) int {
	if end := int(off) + len(p); end > len(f.data) {
		f.data = append(f.data, make([]byte, end-len(f.data))...)
	}
	copy(f.data[off:], p)
	f.pos = off + int64(len(p))
	return len(p)
}

func (f *memFile) WriteAt(p []byte, off int64) (int, error) {
	if f.fs.writeAtErr != nil {
		return 0, f.fs.writeAtErr
	}
	n := f.writeAt(p, off)
	return n, nil
}

func (f *memFile) Seek(offset int64, whence int) (int64, error) {
	if f.fs.seekErr != nil {
		return 0, f.fs.seekErr
	}
	if whence != io.SeekStart {
		return 0, errors.New("memFile: only SeekStart")
	}
	f.pos = offset
	return offset, nil
}

func (f *memFile) Close() error {
	f.closed = true
	return f.fs.closeErr
}

// msgLog records the formatted lines of the run log.
type msgLog struct {
	mu    sync.Mutex
	lines []string
}

func (l *msgLog) add(level, format string, params ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, level+" "+fmt.Sprintf(format, params...))
}

func (l *msgLog) LogInfo(format string, params ...interface{})  { l.add("INFO", format, params...) }
func (l *msgLog) LogError(format string, params ...interface{}) { l.add("ERROR", format, params...) }

func (l *msgLog) text() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := ""
	for _, line := range l.lines {
		out += line + "\n"
	}
	return out
}
