package session

import (
	"bytes"
	"errors"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/internal/testkit"
)

var errDisk = errors.New("disk full")

func startTime() time.Time { return time.Unix(1700000000, 0) }

// memFS is an in-memory FS that can fail on demand.
type memFS struct {
	files      map[string]*memFile
	dirs       []string
	mkdirErr   error
	openErr    func(name string) error
	writeErr   error
	closeCalls int
}

type memFile struct {
	fs *memFS
	bytes.Buffer
	closed bool
}

func newMemFS() *memFS { return &memFS{files: map[string]*memFile{}} }

func (m *memFS) MkdirAll(path string, _ os.FileMode) error {
	m.dirs = append(m.dirs, path)
	return m.mkdirErr
}

func (m *memFS) OpenFile(name string, _ int, _ os.FileMode) (File, error) {
	if m.openErr != nil {
		if err := m.openErr(name); err != nil {
			return nil, err
		}
	}
	f := &memFile{fs: m}
	m.files[name] = f
	return f, nil
}

func (f *memFile) Write(p []byte) (int, error) {
	if f.fs.writeErr != nil {
		return 0, f.fs.writeErr
	}
	return f.Buffer.Write(p)
}

func (f *memFile) Close() error {
	f.closed = true
	f.fs.closeCalls++
	return nil
}

func (m *memFS) content(name string) string {
	if f, ok := m.files[name]; ok {
		return f.String()
	}
	return "<missing>"
}

func newClock() *testkit.FakeClock { return testkit.NewFakeClock(startTime()) }

// eventually yields until cond holds; the deadline only guards against a hang.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		runtime.Gosched()
	}
}

func deps(fs *memFS) Deps { return Deps{Clock: newClock(), FS: fs} }
